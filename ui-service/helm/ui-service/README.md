# UI Service Helm Chart

This Helm chart deploys the KubeQuery UI service, a web interface for logging in and submitting SQL to a target PostgreSQL database.

## Usage

### Install
```sh
helm install my-ui-service ./ui-service/helm/ui-service \
  --set image.repository=your-repo/ui-service \
  --set image.tag=latest
```

### Upgrade
```sh
helm upgrade my-ui-service ./ui-service/helm/ui-service \
  --set image.tag=new-version
```

### Uninstall
```sh
helm uninstall my-ui-service
```

## Configuration
See `values.yaml` for all available options. Key settings:
- `image.repository`, `image.tag`: UI service image
- `replicaCount`: Number of pods
- `service.type`, `service.port`: Service type and port
- `resources`: Pod resource requests/limits

## Example
```yaml
image:
  repository: your-repo/ui-service
  tag: latest
replicaCount: 2
service:
  type: ClusterIP
  port: 80
```

## Notes
- You must configure the image and tag for your registry.
- The UI service listens on port 8080 by default (mapped to service port).