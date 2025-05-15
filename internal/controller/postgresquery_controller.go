/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	kubequeryv1alpha1 "github.com/rsavage/KubeQuery/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"

	"github.com/rsavage/KubeQuery/pkg/db"
)

// PostgresQueryReconciler reconciles a PostgresQuery object
type PostgresQueryReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=kubequery.cloudnexus.io,resources=postgresqueries,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kubequery.cloudnexus.io,resources=postgresqueries/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=kubequery.cloudnexus.io,resources=postgresqueries/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the PostgresQuery object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
func (r *PostgresQueryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	var pq kubequeryv1alpha1.PostgresQuery
	if err := r.Get(ctx, req.NamespacedName, &pq); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// --- Load SQL from Secret or ConfigMap if specified ---
	sql := pq.Spec.SQL
	if pq.Spec.SQLSecretRef != nil {
		var sqlSecret corev1.Secret
		if err := r.Get(ctx, client.ObjectKey{Namespace: pq.Namespace, Name: pq.Spec.SQLSecretRef.Name}, &sqlSecret); err != nil {
			return r.updateStatus(ctx, &pq, false, fmt.Sprintf("failed to get sql secret: %v", err), "", "")
		}
		val, ok := sqlSecret.Data[pq.Spec.SQLSecretRef.Key]
		if !ok {
			return r.updateStatus(ctx, &pq, false, "sql key not found in secret", "", "")
		}
		sql = string(val)
	} else if pq.Spec.SQLConfigMapRef != nil {
		var sqlCM corev1.ConfigMap
		if err := r.Get(ctx, client.ObjectKey{Namespace: pq.Namespace, Name: pq.Spec.SQLConfigMapRef.Name}, &sqlCM); err != nil {
			return r.updateStatus(ctx, &pq, false, fmt.Sprintf("failed to get sql configmap: %v", err), "", "")
		}
		val, ok := sqlCM.Data[pq.Spec.SQLConfigMapRef.Key]
		if !ok {
			return r.updateStatus(ctx, &pq, false, "sql key not found in configmap", "", "")
		}
		sql = val
	}

	// Compute idempotency hash (use loaded SQL)
	hash := sha256.New()
	hash.Write([]byte(fmt.Sprintf("%s|%d|%s|%s|%s", pq.Spec.Connection.Host, pq.Spec.Connection.Port, pq.Spec.Connection.Database, pq.Spec.Connection.User, sql)))
	idempotencyHash := hex.EncodeToString(hash.Sum(nil))

	// If already executed, skip
	if pq.Status.Executed && pq.Status.IdempotencyHash == idempotencyHash {
		log.Info("Query already executed, skipping", "name", pq.Name)
		return ctrl.Result{}, nil
	}

	// Fetch password from secret
	var pwSecret corev1.Secret
	if err := r.Get(ctx, client.ObjectKey{Namespace: pq.Namespace, Name: pq.Spec.Connection.PasswordSecretRef.Name}, &pwSecret); err != nil {
		return r.updateStatus(ctx, &pq, false, fmt.Sprintf("failed to get password secret: %v", err), "", idempotencyHash)
	}
	password, ok := pwSecret.Data[pq.Spec.Connection.PasswordSecretRef.Key]
	if !ok {
		return r.updateStatus(ctx, &pq, false, "password key not found in secret", "", idempotencyHash)
	}

	// Handle SSL config
	var sslCfg *db.SSLConfig
	if pq.Spec.Connection.SSL != nil && pq.Spec.Connection.SSL.Mode != "disable" {
		sslCfg = &db.SSLConfig{Mode: pq.Spec.Connection.SSL.Mode}
		if pq.Spec.Connection.SSL.CaSecretRef != nil {
			var caSecret corev1.Secret
			if err := r.Get(ctx, client.ObjectKey{Namespace: pq.Namespace, Name: pq.Spec.Connection.SSL.CaSecretRef.Name}, &caSecret); err != nil {
				return r.updateStatus(ctx, &pq, false, fmt.Sprintf("failed to get CA secret: %v", err), "", idempotencyHash)
			}
			ca, ok := caSecret.Data[pq.Spec.Connection.SSL.CaSecretRef.Key]
			if !ok {
				return r.updateStatus(ctx, &pq, false, "CA key not found in secret", "", idempotencyHash)
			}
			// Write CA to a temp file
			tmpDir := os.TempDir()
			caPath := filepath.Join(tmpDir, fmt.Sprintf("ca-%s.crt", pq.Name))
			if err := os.WriteFile(caPath, ca, 0600); err != nil {
				return r.updateStatus(ctx, &pq, false, fmt.Sprintf("failed to write CA file: %v", err), "", idempotencyHash)
			}
			sslCfg.CAPath = caPath
		}
	}

	// Prepare DB config
	dbCfg := db.ConnConfig{
		Host:     pq.Spec.Connection.Host,
		Port:     pq.Spec.Connection.Port,
		Database: pq.Spec.Connection.Database,
		User:     pq.Spec.Connection.User,
		Password: string(password),
		SSL:      sslCfg,
	}

	// Set timeout
	timeout := 30 * time.Second
	if pq.Spec.Options != nil && pq.Spec.Options.TimeoutSeconds != nil {
		timeout = time.Duration(*pq.Spec.Options.TimeoutSeconds) * time.Second
	}
	ctxTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	pool, err := db.Connect(ctxTimeout, dbCfg)
	if err != nil {
		return r.updateStatus(ctx, &pq, false, fmt.Sprintf("db connect error: %v", err), "", idempotencyHash)
	}
	defer pool.Close()

	result, err := db.ExecSQL(ctxTimeout, pool, sql)
	if err != nil {
		return r.updateStatus(ctx, &pq, false, fmt.Sprintf("sql exec error: %v", err), "", idempotencyHash)
	}

	return r.updateStatus(ctx, &pq, true, "", result, idempotencyHash)
}

// updateStatus updates the CR status and returns a reconcile result.
func (r *PostgresQueryReconciler) updateStatus(ctx context.Context, pq *kubequeryv1alpha1.PostgresQuery, executed bool, errMsg, result, hash string) (ctrl.Result, error) {
	pq.Status.Executed = executed
	pq.Status.Error = errMsg
	pq.Status.Result = result
	pq.Status.IdempotencyHash = hash
	if err := r.Status().Update(ctx, pq); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PostgresQueryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kubequeryv1alpha1.PostgresQuery{}).
		Named("postgresquery").
		Complete(r)
}
