# KubeQuery: Kubernetes Controller for One-Time PostgreSQL SQL Execution

## Overview
KubeQuery is a Kubernetes-native controller (written in Go) that enables declarative, auditable, and idempotent execution of SQL statements against remote PostgreSQL databases. It leverages Kubernetes Custom Resource Definitions (CRDs) to let you define SQL jobs as YAML manifests, with strong security, flexible connection options, and robust status reporting. Each SQL statement is guaranteed to run only once per unique target and statement, making it ideal for schema migrations, data corrections, or one-off administrative tasks.

---

## Architecture Diagram
```
+-------------------+         +-------------------+         +-------------------+
|                   |  CRD    |                   |  SQL    |                   |
|  User/CI Pipeline +-------->+  KubeQuery        +-------->+  PostgreSQL DB    |
|  (kubectl apply)  |         |  Controller       |         |  (remote, secure) |
+-------------------+         +-------------------+         +-------------------+
                                 |         ^
                                 | Status  |
                                 +---------+
```
- **User/CI Pipeline:** Applies a PostgresQuery CR to the cluster.
- **KubeQuery Controller:** Watches for CRs, fetches secrets, executes SQL, updates status.
- **PostgreSQL DB:** Target for secure, one-time SQL execution.

---

## Key Features
- **Declarative SQL Execution:** Define SQL jobs as Kubernetes resources.
- **Idempotency:** Each SQL statement is executed only once per unique hash (target + SQL).
- **Secure Credentials:** Uses Kubernetes Secrets for passwords and CA certificates.
- **SSL/TLS Support:** Configurable SSL modes and CA trust.
- **Status Reporting:** CR status updated with execution result, error, and idempotency hash.
- **Timeouts:** Configurable query execution timeout.
- **Extensible:** Designed for future support of other databases and advanced workflows.

---

## Use Cases
- **Schema Migrations:** Apply one-time DDL changes (e.g., add columns, indexes) as part of GitOps workflows.
- **Data Fixes:** Run corrective DML (e.g., update, delete) in a controlled, auditable way.
- **Bootstrap/Seed Data:** Insert initial data into a database during environment setup.
- **Automated Remediation:** Trigger SQL fixes in response to monitoring or policy events.
- **Compliance & Audit:** Ensure that all DB changes are tracked, reviewed, and never repeated unintentionally.

---

## How It Works
1. **Apply a PostgresQuery CR:**
   - The controller watches for new/updated CRs.
2. **Idempotency Check:**
   - Computes a hash of the SQL and connection info. If already executed, skips.
3. **Secret Fetch:**
   - Reads password and (optionally) CA cert from referenced secrets.
4. **Connect & Execute:**
   - Connects to PostgreSQL with SSL/TLS as configured, executes the SQL.
5. **Status Update:**
   - Updates CR status with result, error, and idempotency hash.

---

## Example: Basic SQL Migration
```yaml
apiVersion: kubequery.cloudnexus.io/v1alpha1
kind: PostgresQuery
metadata:
  name: add-last-login-column
spec:
  connection:
    host: mydb.example.com
    port: 5432
    database: mydb
    user: myuser
    passwordSecretRef:
      name: mydb-secret
      key: password
    ssl:
      mode: require
      caSecretRef:
        name: mydb-ca
        key: ca.crt
  sql: |
    ALTER TABLE users ADD COLUMN last_login TIMESTAMP;
  options:
    timeoutSeconds: 30
```

### Secret Example (Password)
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: mydb-secret
stringData:
  password: "supersecretpassword"
```

### Secret Example (CA Certificate)
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: mydb-ca
stringData:
  ca.crt: |
    -----BEGIN CERTIFICATE-----
    ...
    -----END CERTIFICATE-----
```

---

## Example: Data Correction (DML)
```yaml
apiVersion: kubequery.cloudnexus.io/v1alpha1
kind: PostgresQuery
metadata:
  name: fix-user-emails
spec:
  connection:
    host: mydb.example.com
    port: 5432
    database: mydb
    user: myuser
    passwordSecretRef:
      name: mydb-secret
      key: password
    ssl:
      mode: require
      caSecretRef:
        name: mydb-ca
        key: ca.crt
  sql: |
    UPDATE users SET email = LOWER(email) WHERE email LIKE '%@EXAMPLE.COM';
  options:
    timeoutSeconds: 10
```

---

## Example: Error Handling and Status
After applying a CR, check its status:
```shell
kubectl get postgresquery add-last-login-column -o yaml
```
Example status block:
```yaml
status:
  executed: true
  error: ""
  result: "ALTER TABLE 1"
  idempotencyHash: "a1b2c3..."
```
If an error occurs (e.g., SQL syntax error, connection failure), the `error` field will be populated and `executed` will be `false`.

---

## CRD Field Reference
| Field | Description | Required |
|-------|-------------|----------|
| `spec.connection.host` | PostgreSQL server hostname or IP | Yes |
| `spec.connection.port` | PostgreSQL server port | Yes |
| `spec.connection.database` | Target database name | Yes |
| `spec.connection.user` | Database username | Yes |
| `spec.connection.passwordSecretRef.name` | Name of secret with password | Yes |
| `spec.connection.passwordSecretRef.key` | Key in secret for password | Yes |
| `spec.connection.ssl.mode` | SSL mode (`disable`, `require`, `verify-ca`, `verify-full`) | Yes |
| `spec.connection.ssl.caSecretRef.name` | Name of secret with CA cert | No (if not verifying CA) |
| `spec.connection.ssl.caSecretRef.key` | Key in secret for CA cert | No |
| `spec.sql` | SQL statement to execute | Yes |
| `spec.options.timeoutSeconds` | Query timeout in seconds | No (default: 30) |

---

## Step-by-Step Usage
1. **Deploy the Controller:**
   - Build and deploy the controller to your cluster (see below).
2. **Create Secrets:**
   - Store your DB password and (optionally) CA cert in Kubernetes secrets.
3. **Apply a PostgresQuery CR:**
   - Write a manifest as shown above and apply it with `kubectl apply -f ...`.
4. **Monitor Execution:**
   - Check the CR status for results, errors, and idempotency hash.
5. **Auditing:**
   - All executions are tracked in the CR status for compliance and review.

---

## Advanced Usage
- **Multiple Environments:** Use labels, namespaces, or workspaces to separate dev/staging/prod jobs.
- **GitOps Integration:** Store CRs in Git, use PRs for review, and automate applies via CI/CD.
- **Chained Migrations:** Apply multiple PostgresQuery CRs in order for complex migrations.
- **Conditional Execution:** Use Kubernetes tools (e.g., Kustomize, ArgoCD) to control when CRs are applied.
- **Secrets Management:** Integrate with external secret managers (e.g., AWS Secrets Manager) via Kubernetes external secrets controllers.

---

## Security & Best Practices
- **Never hardcode credentials:** Always use Kubernetes Secrets.
- **Enable SSL/TLS:** Use `ssl.mode: require` or stricter, and provide a CA if needed.
- **RBAC:** Restrict controller permissions to only required namespaces/secrets.
- **Sensitive Data:** Do not log SQL or credentials.
- **Audit:** Use CR status and Git history for full audit trails.
- **Least Privilege:** Grant DB users only the permissions needed for the intended SQL.

---

## Troubleshooting
- **Query Not Executed:**
  - Check `status.error` for details.
  - Ensure secrets and CA are present and referenced correctly.
  - Confirm network access to the database from the controller pod.
- **SQL Executed More Than Once:**
  - The controller uses a hash of SQL and connection info for idempotency. If you change the SQL or connection, a new execution will occur.
- **Timeouts:**
  - Increase `timeoutSeconds` if your query is long-running.
- **Permissions:**
  - Ensure the controller has RBAC to read secrets and update CR status.

---

## CI/CD Integration & Quality Gates
KubeQuery is designed for GitOps and CI/CD workflows. Recommended practices:
- **Linting & Security:**
  - Use [MegaLinter](https://nvuillam.github.io/mega-linter/) for code, YAML, and Dockerfile linting.
  - Use [KICS](https://kics.io/) and [TFLint](https://github.com/terraform-linters/tflint) for IaC security and style checks.
- **GitHub Actions Example:**
  ```yaml
  name: KubeQuery CI
  on: [push, pull_request]
  jobs:
    lint:
      runs-on: ubuntu-latest
      steps:
        - uses: actions/checkout@v3
        - uses: nvuillam/mega-linter-runner@v6
    test:
      runs-on: ubuntu-latest
      steps:
        - uses: actions/checkout@v3
        - name: Run Go tests
          run: make test
    validate-manifests:
      runs-on: ubuntu-latest
      steps:
        - uses: actions/checkout@v3
        - name: KICS scan
          uses: checkmarx/kics-action@v1
          with:
            path: config/
        - name: TFLint
          run: tflint --init && tflint
  ```
- **Automated Plan/Apply:**
  - Use `kubectl diff` and `kubectl apply` in your pipeline for safe, auditable changes.

---

## Production Best Practices
- **Use remote state and secrets management for all sensitive data.**
- **Enable RBAC and restrict controller permissions to only what is needed.**
- **Monitor controller logs and CR status for errors.**
- **Tag all resources for cost, ownership, and lifecycle tracking.**
- **Automate compliance checks using CI/CD.**
- **Test in staging before production.**

---

## FAQ
**Q: Can I use KubeQuery for MySQL or other databases?**
A: Not yet, but the architecture is modular and support for other databases is planned.

**Q: What happens if I change the SQL or connection info in a CR?**
A: The idempotency hash will change, and the new SQL will be executed once. The old execution will not be repeated.

**Q: Can I run multiple queries in one CR?**
A: You can use a transaction block in the `sql` field, but for best auditability, use one CR per logical change.

**Q: How do I roll back a change?**
A: You must create a new CR with the appropriate rollback SQL. KubeQuery does not automatically revert changes.

**Q: Is the SQL output stored?**
A: Only the command tag/result is stored in the CR status. For full output, use logging or audit tables in your DB.

**Q: How do I restrict who can create PostgresQuery CRs?**
A: Use Kubernetes RBAC to control access to the CRD.

---

## Contributing
We welcome contributions! To get started:
1. Fork the repo and create a feature branch.
2. Run `make generate && make manifests` after editing API types.
3. Add or update tests in `internal/controller/`.
4. Run `make test` and ensure all checks pass.
5. Open a pull request with a clear description.

Please follow our [Code of Conduct](CODE_OF_CONDUCT.md) and ensure your code passes all CI checks.

---

## Roadmap
- [ ] Support for additional databases (e.g., MySQL, SQL Server)
- [ ] Dry-run and preview mode
- [ ] Templated SQL with variable substitution
- [ ] Audit log integration (e.g., Datadog, CloudWatch)
- [ ] Webhook/event triggers
- [ ] Advanced scheduling and dependency management
- [ ] More granular status reporting and metrics

---

## License
Apache 2.0

## Example: Large SQL Scripts (ConfigMap/Secret)
KubeQuery supports very large SQL scripts by referencing a ConfigMap or Secret. This is recommended for scripts that are hundreds or thousands of lines long.

### Using a ConfigMap
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-sql-script
  namespace: default
data:
  script.sql: |
    -- Hundreds of lines of SQL here
    CREATE TABLE big_table (...);
    ...
```

```yaml
apiVersion: kubequery.cloudnexus.io/v1alpha1
kind: PostgresQuery
metadata:
  name: configmap-sql-example
spec:
  connection:
    host: mydb.example.com
    port: 5432
    database: mydb
    user: myuser
    passwordSecretRef:
      name: mydb-secret
      key: password
    ssl:
      mode: require
      caSecretRef:
        name: mydb-ca
        key: ca.crt
  sqlConfigMapRef:
    name: my-sql-script
    key: script.sql
  options:
    timeoutSeconds: 120
```

### Using a Secret
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: my-sql-secret
  namespace: default
stringData:
  script.sql: |
    -- Sensitive or proprietary SQL
    ...
```

```yaml
apiVersion: kubequery.cloudnexus.io/v1alpha1
kind: PostgresQuery
metadata:
  name: secret-sql-example
spec:
  connection:
    host: mydb.example.com
    port: 5432
    database: mydb
    user: myuser
    passwordSecretRef:
      name: mydb-secret
      key: password
    ssl:
      mode: require
      caSecretRef:
        name: mydb-ca
        key: ca.crt
  sqlSecretRef:
    name: my-sql-secret
    key: script.sql
  options:
    timeoutSeconds: 120
```

### Precedence
- If `sqlSecretRef` is set, it is used.
- Else if `sqlConfigMapRef` is set, it is used.
- Else, the inline `sql` field is used.

**This allows you to manage very large or sensitive SQL scripts outside the CR, keeping manifests clean and secure.**

---

## Helm Chart

KubeQuery comes with a fully supported Helm chart for easy installation and management in Kubernetes clusters. The chart is located in [`helm/kubequery`](./helm/kubequery).

- **Recommended installation method:** The Helm chart is the preferred way to deploy KubeQuery, as it manages all resources, RBAC, and CRDs in a repeatable, configurable manner.
- **Chart documentation:** See [`helm/kubequery/README.md`](./helm/kubequery/README.md) for installation, upgrade, and configuration instructions.
- **Quick start:**
  ```sh
  helm install my-kubequery ./helm/kubequery \
    --set image.repository=your-repo/kubequery \
    --set image.tag=latest
  ```
