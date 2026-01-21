# Adding a Service to My Cluster

To deploy a new service to my k8s cluster, create the following files in `cluster-config/apps/<service-name>/`:

## deployment.yaml

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: <service-name>
  namespace: <service-name>
spec:
  replicas: 1
  selector:
    matchLabels:
      app: <service-name>
  template:
    metadata:
      labels:
        app: <service-name>
    spec:
      containers:
        - name: <service-name>
          image: ghcr.io/andresfelipemendez/<service-name>:latest
          ports:
            - containerPort: 80
          resources:
            limits:
              memory: "64Mi"
              cpu: "50m"
```

## service.yaml

```yaml
apiVersion: v1
kind: Service
metadata:
  name: <service-name>
  namespace: <service-name>
spec:
  selector:
    app: <service-name>
  ports:
    - port: 80
      targetPort: 80
```

## ingress.yaml

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: <service-name>-ingress
  namespace: <service-name>
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  ingressClassName: nginx
  tls:
    - hosts:
        - <service-name>.andresfelipemendez.com
      secretName: <service-name>-tls
  rules:
    - host: <service-name>.andresfelipemendez.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: <service-name>
                port:
                  number: 80
```

## kustomization.yaml

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: <service-name>
resources:
  - deployment.yaml
  - service.yaml
  - ingress.yaml
```

## ArgoCD App

Create `cluster-config/apps/argocd/<service-name>-app.yaml`:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: <service-name>
  namespace: argocd
  annotations:
    argocd-image-updater.argoproj.io/image-list: app=ghcr.io/andresfelipemendez/<service-name>:latest
    argocd-image-updater.argoproj.io/app.update-strategy: digest
    argocd-image-updater.argoproj.io/write-back-method: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/andresfelipemendez/cluster-config.git
    targetRevision: HEAD
    path: apps/<service-name>
  destination:
    server: https://kubernetes.default.svc
    namespace: <service-name>
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - CreateNamespace=true
```

Then add the app to `cluster-config/apps/argocd/kustomization.yaml`:

```yaml
resources:
  - <service-name>-app.yaml
```

Push the changes and ArgoCD will automatically deploy the service.
