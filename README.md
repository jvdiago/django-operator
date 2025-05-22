# django-operator
This operator manages common Django administration tasks directly from Kubernetes Custom Resources. It supports:

* **User management**: create superusers or staff users via `DjangoUser` CRs, with credentials stored in Kubernetes Secrets.
* **Database migrations**: run `manage.py migrate` (optionally per-app or per-migration) via `DjangoMigrate` CRs.
* **Static file collection**: run `manage.py collectstatic` via `DjangoStatic` CRs.
* **Celery control**: manage Celery workers, revoke tasks, and flush queues via `DjangoCelery` CRs.

## Namespaced Operator

In order to run the commands the operator needs exec privileges on pods. Because of the possible security issue that this might create, the operator is restricted to run on the specific namespace where Django is running. Multiple operator are needed for multiple Django deployments in multiple namespaces

## Configuration

The operator currently needs two ENV variables to be configured to be able to find the Django and Celery pods. They are defined in config/manager.manager.yaml and need to be tailored to your tags to be able to find the pods
```       - name: DJANGO_POD_LABEL
            value: "app.kubernetes.io/component:django-server"
          - name: CELERY_POD_LABEL
            value: "app.kubernetes.io/component:django-celery-work-celery"
```
## Usage Examples

Below are YAML snippets for each CR type.

### 1. Create a Django User (`DjangoUser`)

**Spec**:

```yaml
apiVersion: django.my.domain/v1alpha1
kind: DjangoUser
metadata:
  name: admin-user
  namespace: django-operator
spec:
  username: admin
  email: admin@example.com
  superuser: true
  passwordSecretRef:
    name: admin-password-secret
    key: password
```

**Secret** (store password):

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: admin-password-secret
  namespace: django-operator
stringData:
  password: S3cr3tP@ssw0rd
```

After applying both, the operator will exec into the Django pod and create/update the user, setting `.status.created`.

### 2. Run Database Migrations (`DjangoMigrate`)

**Spec**:

```yaml
apiVersion: django.my.domain/v1alpha1
kind: DjangoMigrate
metadata:
  name: migrate-all
  namespace: django-operator
spec:
  fake: false          # run fake migrations if true
  app: myapp                 # optional: run only this app's migrations
  migration:            # optional: target a specific migration
```

Applying this CR runs `python manage.py migrate` inside the Django pod and records `.status.applied`.

### 3. Collect Static Files (`DjangoStatic`)

**Spec**:

```yaml
apiVersion: django.my.domain/v1alpha1
kind: DjangoStatic
metadata:
  name: collect-static
  namespace: django-operator
spec: {}
```

This triggers `python manage.py collectstatic --noinput` and sets `.status.collected`.

### 4. Control Celery (`DjangoCelery`)

**Spec**:

```yaml
apiVersion: django.my.domain/v1alpha1
kind: DjangoCelery
metadata:
  name: flush-all-queues
  namespace: django-operator
spec:
  app: myapp          # the Celery app name
  worker: worker1     # optional: target a specific worker
  task:                # optional: revoke a specific task ID
```

* **Flush all queues**: omit `worker` and `task`, the operator runs `celery -A {{app}} purge -f`.
* **Flush a worker**: set `worker`, runs `celery -A {{app}} purge -f -Q {{worker}}`.
* **Revoke a task**: set `task`, runs `celery -A {{app}} control revoke {{task}}`.

After execution, `.status.executed` is updated.
## Installation
1. **Install CRDs**

   ```bash
   kubectl apply -k config/crd
   ```

2. **Deploy the operator** (In the namespace were Django containers are running)

   ```bash
   # Install RBAC and Deployment
   kubectl apply -k config/rbac -n django-operator
   kubectl apply -k config/default -n django-operator
   ```
## Uninstallation

To remove the operator and CRDs:

```bash
kubectl delete -k config/default -n django-operator
kubectl delete -k config/crd
kubectl delete namespace django-operator
```
## Getting Started

### Prerequisites
- go version v1.23.0+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

### To Deploy on the cluster
**Build and push your image to the location specified by `IMG`:**

```sh
make docker-build docker-push IMG=<some-registry>/django-operator:tag
```

**NOTE:** This image ought to be published in the personal registry you specified.
And it is required to have access to pull the image from the working environment.
Make sure you have the proper permission to the registry if the above commands don’t work.

**Install the CRDs into the cluster:**

```sh
make install
```

**Deploy the Manager to the cluster with the image specified by `IMG`:**

```sh
make deploy IMG=<some-registry>/django-operator:tag NAMESPACE=<my-namespace>
```

> **NOTE**: If you encounter RBAC errors, you may need to grant yourself cluster-admin
privileges or be logged in as admin.

**Create instances of your solution**
You can apply the samples (examples) from the config/sample:

```sh
kubectl apply -k config/samples/
```

>**NOTE**: Ensure that the samples has default values to test it out.

### To Uninstall
**Delete the instances (CRs) from the cluster:**

```sh
kubectl delete -k config/samples/
```

**Delete the APIs(CRDs) from the cluster:**

```sh
make uninstall
```

**UnDeploy the controller from the cluster:**

```sh
make undeploy
```
## License

Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

