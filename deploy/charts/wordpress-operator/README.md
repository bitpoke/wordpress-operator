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
| `resources`                     | Controller container resources limits and requests                                            | `{}`                                                    |
| `nodeSelector`                  | Controller pod nodeSelector                                                                   | `{}`                                                    |
| `tolerations`                   | Controller pod tolerations                                                                    | `{}`                                                    |
| `affinity`                      | Controller pod node affinity                                                                  | `{}`                                                    |
| `extraArgs`                     | Args that are passed to controller, check controller command line flags                       | `[]`                                                    |
| `extraEnv`                      | Extra environment vars that are passed to controller, check controller command line flags     | `{}`                                                    |
| `rbac.create`                   | Whether or not to create rbac service account, role and roleBinding                           | `true`                                                  |
