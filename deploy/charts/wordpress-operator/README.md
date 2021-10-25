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
| `image`                         | Controller container image                                                                    | `docker.io/bitpoke/wordpress-operator:latest`           |
| `imagePullPolicy`               | Controller image pull policy                                                                  | `IfNotPresent`                                          |
| `imagePullSecrets`              | Controller image pull secret                                                                  |                                                         |
| `installCRDs`                   | Whether or not to install CRDS, Regardless of value of this, Helm v3+ will install the CRDs if those are not present already. Use `--skip-crds` with `helm install` if you want to skip CRD creation                                                                | `true`                                    |
| `resources`                     | Controller container resources limits and requests                                            | `{}`                                                    |
| `nodeSelector`                  | Controller pod nodeSelector                                                                   | `{}`                                                    |
| `tolerations`                   | Controller pod tolerations                                                                    | `{}`                                                    |
| `affinity`                      | Controller pod node affinity                                                                  | `{}`                                                    |
| `extraArgs`                     | Args that are passed to controller, check controller command line flags                       | `[]`                                                    |
| `extraEnv`                      | Extra environment vars that are passed to controller, check controller command line flags     | `{}`                                                    |
| `rbac.create`                   | Whether or not to create rbac service account, role and roleBinding                           | `true`                                                  |
