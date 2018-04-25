# wordpress-operator
WordPress operator for Kubernetes

# Kubernetes resources

## Wordpress Site

```yaml
apiVersion: wordpress.presslabs.net/v1
kind: Wordpress
metadata:
  name: mysite
  labels:
    wordpress-runtime: stable
    env: production
    flavor: scalable
  annotations:
    # *.provisioner.presslabs.com annotations drives provisioning under Presslabs Dashboard
    mysql.provisioner.presslabs.com/class: "mysql-operator"
    mysql.provisioner.presslabs.com/secret-name: "mysite-mysql"
    # if this mysql cluster exists in namespace, skip provisioning
    # the Site Controller sets the env accoding to this instance
    mysql.provisioner.presslabs.com/instance: "mysite-mysql-ru5zti"

    memcached.provisioner.presslabs.com/class: "inline"
    # if this memcached statefulset exists in namespace, skip provisioning
    # the Site Controller sets the env accoding to this instance
    memcached.provisioner.presslabs.com/instance: "mysite-memcached-pprfg8"

    git.provisioner.presslabs.com/class: "project-gitea"
    git.provisioner.presslabs.com/repo: "octocat/mysite"

    media.provisioner.presslabs.com/class: "gcs"
    media.provisioner.presslabs.com/bucket: "mysite-files"
spec:
  image: quay.io/presslabs/wordpress:4.9.5-r148-php71
  replicas: 1
  domains:
    - "example.com"
    - "www.example.com"
  tlsSecretRef: mysite-tls
  repoURL: "https://github.com/octocat/Hello-World.git"
  repoRef: "master"
  readOnlyContent: true
  keepUploadsLocal: false # if enabled, mount RW `/www/wp-content/uploads` from PVC
  secretRef: mysite  # keys
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
  rollingUpdate:
    maxUnavailable: 25%
    maxSurge: 25%
  persistentVolumeTemplate: {}  # a PVC template for cloning. (defaults to emptyDir)
  serviceSpec: {}
  ingressAnnotations: {}
status:
  conditions:
    - type: Ready
      status: True
      reason: StatefulSetReady
      message: The statefulset transitioned to Ready
```
