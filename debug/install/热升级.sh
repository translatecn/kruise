kubectl apply -f - << EOF
---
apiVersion: apps.kruise.io/v1alpha1
kind: SidecarSet
metadata:
  name: guestbook-sidecar
spec:
  selector: # select the pods to be injected with sidecar containers
    matchLabels:
      app.kubernetes.io/name: guestbook-with-sidecar
  containers:
    - name: guestbook-sidecar
      image: registry.cn-hangzhou.aliyuncs.com/acejilam/guestbook:sidecar
      imagePullPolicy: Always
      ports:
        - name: sidecar-server
          containerPort: 4000 # different from main guestbook containerPort which is 3000
      volumeMounts:
        - name: log-volume
          mountPath: /var/log
  volumes:
    - name: log-volume
      emptyDir: {}
EOF

kubectl apply -f - << EOF
---
apiVersion: apps.kruise.io/v1alpha1
kind: StatefulSet
metadata:
  name: guestbook-with-sidecar
  labels:
    app: guestbook
    version: "1.0"
spec:
  replicas: 10
  serviceName: guestbook-with-sidecar
  selector:
    matchLabels:
      app.kubernetes.io/name:  guestbook-with-sidecar
  template:
    metadata:
      labels:
        app.kubernetes.io/name:  guestbook-with-sidecar
        version: "1.0"
    spec:
      readinessGates:
        # A new condition that ensures the pod remains at NotReady state while the in-place update is happening
        - conditionType: InPlaceUpdateReady
      containers:
      - name: guestbook
        image: registry.cn-hangzhou.aliyuncs.com/acejilam/guestbook:v1
        ports:
        - name: http-server
          containerPort: 3000
  podManagementPolicy: Parallel  # allow parallel updates, works together with maxUnavailable
  updateStrategy:
    type: RollingUpdate
    rollingUpdate:
      # Do in-place update if possible, currently only image update is supported for in-place update
      podUpdatePolicy: InPlaceIfPossible
      # Allow parallel updates with max number of unavailable instances equals to 2
      maxUnavailable: 3
---
apiVersion: v1
kind: Service
metadata:
  name: guestbook-with-sidecar
  labels:
    app: guestbook-with-sidecar
spec:
  ports:
  - nodePort: 30163
    port: 3000
    targetPort: http-server
    name: main-port
  - nodePort: 30164
    port: 4000
    targetPort: sidecar-server
    name: sidecar-port
  selector:
    app.kubernetes.io/name: guestbook-with-sidecar
  type: NodePort
EOF





# 修改 sts.apps.kruise.io/guestbook-with-sidecar  image=registry.cn-hangzhou.aliyuncs.com/acejilam/guestbook:sidecar-v2
