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
  replicas: 1
  webrootVolumeSpec:
    # by default, it gets mounted into /var/www/html
    # one of the following:

    # persistentVolumeClaim: {}
    # hostPath: {}
    # emptyDir: {} (default)

  mediaVolumeSpec:
    # if it's defined, by default, it's root gets mounted into /var/www/html/wp-content/uploads
    # one of the following:

    # persistentVolumeClaim: {}
    # hostPath: {}
    # emptyDir: {}

  volumeMounts: []
    # overrides default mounts for webrootVolumeSpec and mediaVolumeSpec

  env:
      # gets injected into every container and initContainer in
      # webPodTemplate and cliPodTemplate
    - name: WORDPRESS_DB_PASSWORD
      valueFrom:
        secretKeyRef: mysite-mysql
        key: PASSWORD
  envFrom:
      # gets injected into every container and initContainer in
      # webPodTemplate and cliPodTemplate
    - prefix: "WORDPRESS_"
      secretRef:
        name: mysite-salt
  webPodTemplate: {}
  cliPodTemplate: {}
  serviceSpec: {}
  domains:
    - "example.com"
    - "www.example.com"
  tlsSecretRef: mysite-tls
  ingressAnnotations: {}
```

## License

This project is licensed under Apache 2.0 license. Read the [LICENSE](LICENSE) file in the
top distribution directory, for the full license text.
