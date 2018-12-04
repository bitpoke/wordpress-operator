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

## Controller deploy

Install helm. New to helm? Check https://github.com/helm/helm#install 

Install kubectl. For more details, see: https://kubernetes.io/docs/tasks/tools/install-kubectl/ 

To deploy this controller, use the provided helm chart, by running:
```shell
helm repo add presslabs https://presslabs.github.io/charts
helm install presslabs/wordpress-operator --name wordpress-operator
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
  # image: quay.io/presslabs/wordpress-runtime
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
    # persistentVolumeClaim: {}
    # hostPath: {}
    # emptyDir: {}

  # extra volumes for the Wordpress container
  volumes: []
  # extra volume mounts for the Wordpress container
  volumeMounts: []
  # extra env variables for the Wordpress container
  env:
    - name: WORDPRESS_DB_HOST
      value: mysite-mysql
    - name: WORDPRESS_DB_PASSWORD
      valueFrom:
        secretKeyRef: mysite-mysql
        key: PASSWORD
  envFrom: []

  # secret containg HTTPS certificate
  tlsSecretRef: mysite-tls
  # extra ingress annotations
  ingressAnnotations: {}
```

## License

This project is licensed under Apache 2.0 license. Read the [LICENSE](LICENSE) file in the
top distribution directory, for the full license text.
