# Bitpoke WordPress Operator

This is the helm chart for [wordpress-operator](https://github.com/bitpoke/wordpress-operator).

## TL;DR
```sh
helm repo add bitpoke https://helm-charts.bitpoke.io
helm install wordpress-operator bitpoke/wordpress-operator
```

## Configuration
The following table contains the configuration parameters for wordpress-operator and default values.

| Parameter                       | Description                                                                                   | Default value                                           |
| ---                             | ---                                                                                           | ---                                                     |
| `replicas`                      | Replicas for controller                                                                       | `1`                                                     |
| `image.repository`              | Controller image repository                                                                   | `docker.io/bitpoke/wordpress-operator`                  |
| `image.pullPolicy`              | Controller image pull policy                                                                  | `IfNotPresent`                                          |
| `image.tag       `              | Controller image tag                                                                          | `latest`                                                |
| `imagePullSecrets`              | Controller image pull secret                                                                  |                                                         |
| `podAnnotations`                | Extra pod annotations                                                                         | `{}`                                                    |
| `podSecurityContext`            | The pod security context. `65532` is the UID/GID for the nonroot user in the official images  | `{runAsNonRoot: true, runAsUser: 65532, runAsGroup: 65532, fsGroup: 65532}` |
| `securityContext`               | Security context for the WordPress Operator container                                         | `{}`                                                    |
| `resources`                     | Controller container resources limits and requests                                            | `{}`                                                    |
| `nodeSelector`                  | Controller pod nodeSelector                                                                   | `{}`                                                    |
| `tolerations`                   | Controller pod tolerations                                                                    | `{}`                                                    |
| `affinity`                      | Controller pod node affinity                                                                  | `{}`                                                    |
| `extraArgs`                     | Args that are passed to controller, check controller command line flags                       | `[]`                                                    |
| `extraEnv`                      | Extra environment vars that are passed to controller, check controller command line flags     | `{}`                                                    |
| `rbac.create`                   | Whether or not to create rbac service account, role and roleBinding                           | `true`                                                  |
| `serviceAccount.create`         | Specifies whether a service account should be created                                         | `true`                                                  |
| `serviceAccount.annotations`    | Annotations to add to the service account                                                     | `{}`                                                    |
| `serviceAccount.name`           | The name of the service account to use. If not set and create is true, a name is generated using the fullname template. | `empty`                       |
