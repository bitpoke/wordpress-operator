# wordpress-operator
WordPress operator enables managing multiple WordPress installments at scale.

## Goals and status

The main goals of the operator are:

1. Easily deploy scalable WordPress sites on top of kubernetes
2. Allow best practices for en masse upgrades (canary, slow rollout, etc.)
3. Friendly to devops (monitoring, availability, scalability and backup stories solved)

The project is in pre-alpha state.

## Components

1. WordPress operator - this project
2. WordPress runtime - container image supporting the project goals (https://github.com/presslabs/runtime)

# Kubernetes resources

## Wordpress Site

```yaml
apiVersion: wordpress.presslabs.org/v1alpha1
kind: Wordpress
metadata:
  name: mysite
spec:
  image: quay.io/presslabs/wordpress:4.9.5-r148-php71
  replicas: 1
  domains:
    - "example.com"
    - "www.example.com"
  tlsSecretRef: mysite-tls
  contentVolumeSpec:
    # readOnly: true
    # one of the following
    # persistentVolumeClaim: {}
    # hostPath: {}
    # emptyDir: {}
  mediaVolumeSpec:
    # readOnly: true
    # one of the following
    # persistentVolumeClaim: {}
    # hostPath: {}
    # emptyDir: {}
  secretRef: mysite
    # wp-config.php
    # php.ini
    # nginx-vhost.conf
    # nginx-server.conf
    # id_rsa
    # netrc
    # service_account.json
    # aws_credentials
    # aws_config
  env:
    - name: WORDPRESS_DB_PASSWORD
      valueFrom:
        secretKeyRef: mysite-mysql
        key: PASSWORD
  resources:
    required:
      nginx/cpu: ...
      nginx/memory: ...
      php/cpu: ...
      php/memory: ...
      php/workers: 4
      php/worker-memory: 128Mi
      php/max-execution-seconds: 30
    limits:
      nginx/cpu: ...
      nginx/memory: ...
      php/cpu: ...
      php/memory: ...
      php/workers: 4
      php/worker-memory: 256Mi
      php/max-execution-seconds: 30
      ingress/max-body-size: 100Mi
  nodeSelector: {}
  tolerations: {}
  affinity: {}
  serviceSpec: {}
  ingressAnnotations: {}
```

## License

This project is licensed under Apache 2.0 license. Read the [LICENSE](LICENSE) file in the
top distribution directory, for the full license text.
