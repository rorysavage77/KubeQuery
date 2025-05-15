# KubeQuery Helm Chart

This Helm chart deploys the KubeQuery controller for one-time, idempotent SQL execution against PostgreSQL databases using Kubernetes CRDs.

## Usage

### Install
```sh
helm repo add kubequery https://your-org.github.io/kubequery-helm-charts
helm install my-kubequery kubequery/kubequery \
  --set image.repository=your-repo/kubequery \
  --set image.tag=latest
```

### Upgrade
```sh
helm upgrade my-kubequery kubequery/kubequery \
  --set image.tag=new-version
```

### Uninstall
```sh
helm uninstall my-kubequery
```

## CRDs
By default, the chart installs the PostgresQuery CRD. You can disable this with `--set crds.install=false` if you manage CRDs separately.

## Configuration
See `values.yaml` for all available options. Key settings:
- `image.repository`, `image.tag`: Controller image
- `replicaCount`: Number of controller pods
- `rbac.create`: Create RBAC resources
- `serviceAccount.create`: Create a ServiceAccount
- `crds.install`: Install CRDs
- `resources`: Pod resource requests/limits

## Example
```yaml
image:
  repository: your-repo/kubequery
  tag: latest
replicaCount: 2
rbac:
  create: true
serviceAccount:
  create: true
crds:
  install: true
```

## Notes
- You must configure the controller image and tag for your registry.
- The chart is namespace-scoped by default. For cluster-wide operation, extend RBAC as needed.
- For CRD usage and examples, see the main project README.