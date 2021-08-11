wordpress-operator
[![Build Status](https://bitpoke.cloud/api/badges/bitpoke/wordpress-operator/status.svg)](https://bitpoke.cloud/bitpoke/wordpress-operator)
===
WordPress operator enables managing multiple WordPress installments at scale.



## Goals and status

The main goals of the operator are:

1. Easily deploy scalable WordPress sites on top of kubernetes
2. Allow best practices for en masse upgrades (canary, slow rollout, etc.)
3. Friendly to devops (monitoring, availability, scalability and backup stories solved)

The project is actively developed and maintained and has reached stable beta state. Check [here](https://github.com/bitpoke/wordpress-operator/releases) the project releases.



## Components

1. WordPress operator - this project
2. WordPress runtime - container image supporting the project goals (https://github.com/bitpoke/stack-runtime/tree/master/wordpress)



## Deploy

### Install CRDs

#### This step is optional. By default helm will install CRDs.

Install kustomize. New to kustomize? Check https://kustomize.io/

To install CRDs use the following command:

```shell
kustomize build github.com/bitpoke/wordpress-operator/config | kubectl apply -f-
```


### Install controller

Install helm. New to helm? Check https://github.com/helm/helm#install

Install kubectl. For more details, see: https://kubernetes.io/docs/tasks/tools/install-kubectl/

To deploy this controller, use the provided helm chart, by running:

```shell
helm repo add bitpoke https://helm.bitpoke.cloud/charts
helm install bitpoke/wordpress-operator --name wordpress-operator
# or if using helm v3
helm install wordpress-operator bitpoke/wordpress-operator
```



## Deploying a Wordpress Site

```yaml
apiVersion: wordpress.presslabs.org/v1alpha1
kind: Wordpress
metadata:
  name: mysite
spec:
  replicas: 3
  domains:
    - example.com
  # image: docker.io/bitpoke/wordpress-runtime
  # tag: latest
  code: # where to find the code
    # contentSubpath: wp-content/
    # by default, code get's an empty dir. Can be one of the following:
    git:
      repository: https://github.com/example.com
      # reference: master
      # env:
      #   - name: SSH_RSA_PRIVATE_KEY
      #     valueFrom:
      #       secretKeyRef:
      #         name: mysite
      #         key: id_rsa

    # persistentVolumeClaim: {}
    # hostPath: {}
    # emptyDir: {} (default)

  media: # where to find the media files
    # by default, code get's an empty dir. Can be one of the following:
    gcs: # store files using Google Cloud Storage
      bucket: calins-wordpress-runtime-playground
      prefix: mysite/
      env:
        - name: GOOGLE_CREDENTIALS
          valueFrom:
            secretKeyRef:
              name: mysite
              key: google_application_credentials.json
        - name: GOOGLE_PROJECT_ID
          value: development
    # persistentVolumeClaim: {}
    # hostPath: {}
    # emptyDir: {}
  bootstrap: # wordpress install config
    env:
      - name: WORDPRESS_BOOTSTRAP_USER
        valueFrom:
          secretKeyRef:
            name: mysite
            key: USER
      - name: WORDPRESS_BOOTSTRAP_PASSWORD
        valueFrom:
          secretKeyRef:
            name: mysite
            key: PASSWORD
      - name: WORDPRESS_BOOTSTRAP_EMAIL
        valueFrom:
          secretKeyRef:
            name: mysite
            key: EMAIL
      - name: WORDPRESS_BOOTSTRAP_TITLE
        valueFrom:
          secretKeyRef:
            name: mysite
            key: TITLE
  # extra volumes for the Wordpress container
  volumes: []
  # extra volume mounts for the Wordpress container
  volumeMounts: []
  # extra env variables for the Wordpress container
  env:
    - name: DB_HOST
      value: mysite-mysql
    - name: DB_USER
      valueFrom:
        secretKeyRef: mysite-mysql
        key: USER
    - name: DB_PASSWORD
      valueFrom:
        secretKeyRef: mysite-mysql
        key: PASSWORD
    - name: DB_NAME
      valueFrom:
        secretKeyRef: mysite-mysql
        key: DATABASE
  envFrom: []

  # secret containg HTTPS certificate
  tlsSecretRef: mysite-tls
  # extra ingress annotations
  ingressAnnotations: {}
```



## License

This project is licensed under Apache 2.0 license. Read the [LICENSE](LICENSE) file in the
top distribution directory, for the full license text.
