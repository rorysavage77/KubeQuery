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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PostgresQuerySpec defines the desired state of PostgresQuery.
type PostgresQuerySpec struct {
	// Connection contains the PostgreSQL connection configuration.
	Connection PostgresConnection `json:"connection"`
	// SQL is the SQL statement to execute against the target database.
	// This should be a single statement or a transaction block.
	// If sqlSecretRef or sqlConfigMapRef is set, this field is ignored.
	SQL string `json:"sql,omitempty"`
	// sqlConfigMapRef references a ConfigMap containing the SQL script (optional).
	// If both sqlSecretRef and sqlConfigMapRef are set, sqlSecretRef takes precedence.
	SQLConfigMapRef *ConfigMapKeySelector `json:"sqlConfigMapRef,omitempty"`
	// sqlSecretRef references a Secret containing the SQL script (optional).
	// If set, this takes precedence over sqlConfigMapRef and sql.
	SQLSecretRef *SecretKeySelector `json:"sqlSecretRef,omitempty"`
	// Options for query execution (e.g., timeout).
	Options *QueryOptions `json:"options,omitempty"`
}

// PostgresConnection defines how to connect to the PostgreSQL database.
type PostgresConnection struct {
	// Host is the hostname or IP address of the PostgreSQL server.
	Host string `json:"host"`
	// Port is the port number of the PostgreSQL server.
	Port int `json:"port"`
	// Database is the name of the target database.
	Database string `json:"database"`
	// User is the username for authentication.
	User string `json:"user"`
	// PasswordSecretRef references a Kubernetes Secret for the database password.
	PasswordSecretRef SecretKeySelector `json:"passwordSecretRef"`
	// SSL contains SSL/TLS configuration for the connection.
	SSL *PostgresSSL `json:"ssl,omitempty"`
}

// PostgresSSL defines SSL/TLS settings for PostgreSQL connections.
type PostgresSSL struct {
	// Mode is the SSL mode (disable, require, verify-ca, verify-full).
	Mode string `json:"mode"`
	// CaSecretRef references a Kubernetes Secret for the CA certificate (optional).
	CaSecretRef *SecretKeySelector `json:"caSecretRef,omitempty"`
}

// SecretKeySelector selects a key of a Secret.
type SecretKeySelector struct {
	// Name of the secret.
	Name string `json:"name"`
	// Key within the secret.
	Key string `json:"key"`
}

// ConfigMapKeySelector selects a key of a ConfigMap.
type ConfigMapKeySelector struct {
	// Name of the ConfigMap.
	Name string `json:"name"`
	// Key within the ConfigMap.
	Key string `json:"key"`
}

// QueryOptions defines optional execution parameters.
type QueryOptions struct {
	// TimeoutSeconds is the query execution timeout in seconds.
	TimeoutSeconds *int `json:"timeoutSeconds,omitempty"`
}

// PostgresQueryStatus defines the observed state of PostgresQuery.
type PostgresQueryStatus struct {
	// Executed indicates if the query was executed successfully.
	Executed bool `json:"executed"`
	// Error contains any error message from execution.
	Error string `json:"error,omitempty"`
	// Result contains a summary or result of the execution (if applicable).
	Result string `json:"result,omitempty"`
	// IdempotencyHash is a hash of the SQL and connection info to prevent re-execution.
	IdempotencyHash string `json:"idempotencyHash,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// PostgresQuery is the Schema for the postgresqueries API.
type PostgresQuery struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PostgresQuerySpec   `json:"spec,omitempty"`
	Status PostgresQueryStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PostgresQueryList contains a list of PostgresQuery.
type PostgresQueryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PostgresQuery `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PostgresQuery{}, &PostgresQueryList{})
}
