# On Kubernetes

ResticGUI runs as one pod: it keeps its state in SQLite, on a
PersistentVolumeClaim. The manifests are in
[`deploy/kubernetes`](https://github.com/mathieuHa/resticgui/tree/main/deploy/kubernetes),
ready for `kubectl apply -k`.

## 1. Create the namespace and the key

ResticGUI encrypts the repository passwords and credentials it saves with a
key. Keep it in a Secret, and a copy somewhere safe:

```sh
kubectl create namespace resticgui
kubectl -n resticgui create secret generic resticgui-key \
  --from-literal=key="$(openssl rand -base64 32)"
```

## 2. Apply the manifests

```sh
kubectl apply -k 'https://github.com/mathieuHa/resticgui//deploy/kubernetes?ref=main'
kubectl -n resticgui rollout status deployment/resticgui
```

They create:

<<< @/../deploy/kubernetes/resticgui.yaml

and a ConfigMap from this configuration, mounted at `/opt/config.yaml`:

<<< @/../deploy/kubernetes/config.yaml

## 3. Open it

```sh
kubectl -n resticgui port-forward service/resticgui 3001:80
```

Open <http://127.0.0.1:3001/ui/> and continue with
[your first backup](/guide/first-backup).

## Back up the volumes of an application

Mount the volumes under `/backup`, read-only, in the Deployment. Copy the
folder, change it, and apply your copy:

```sh
git clone https://github.com/mathieuHa/resticgui
cp -r resticgui/deploy/kubernetes my-resticgui
```

```yaml
# my-resticgui/resticgui.yaml, in the container
volumeMounts:
  - {name: db, mountPath: /opt/db}
  - {name: config, mountPath: /opt/config.yaml, subPath: config.yaml, readOnly: true}
  - {name: data, mountPath: /backup/data, readOnly: true}
# and in the pod
volumes:
  - name: data
    persistentVolumeClaim: {claimName: my-app-data}
```

```sh
kubectl apply -k my-resticgui
```

A ReadWriteOnce volume can be mounted by pods on one node at a time: put
ResticGUI on the node of the application, with a `nodeSelector` or pod
affinity, or use a ReadWriteMany volume.

## Ingress and a password

Make a password hash, put it in `config.yaml`, and add `ingress.yaml` to
the resources of `kustomization.yaml`:

```sh
kubectl -n resticgui exec -it deploy/resticgui -- /opt/resticgui hash-password
```

<<< @/../deploy/kubernetes/ingress.yaml

## Updates

The manifests use the `latest` tag. Pin a version in `kustomization.yaml`:

<<< @/../deploy/kubernetes/kustomization.yaml

then `kubectl apply -k` again. The `Recreate` strategy stops the old pod
before the new one opens the database.
