# POC Inject Service Account ID Token for Cloud Run to Cloud Run Calls

```sh
export PROJECT_ID=<>
export REPO_ROOT="$(pwd)"
```

## Build the Sample App and Auth Sidecar

```sh
cd "$REPO_ROOT"/sample-app
gcloud builds submit . --project $PROJECT_ID
```

```sh
cd "$REPO_ROOT"/auth-sidecar
gcloud builds submit . --project $PROJECT_ID
```

## Deploy a Backend Service

```sh
gcloud run deploy backend-service --image europe-west1-docker.pkg.dev/$PROJECT_ID/demo-repo/sample-app --region europe-west1 --no-allow-unauthenticated --project $PROJECT_ID
```

validate the backend service runs

```sh
curl "$(gcloud run services describe backend-service --region=europe-west1 --project=$PROJECT_ID --format="value(status.url)")/test" -H "Authorization: Bearer $(gcloud auth print-identity-token)"
```

This should return

```txt
hello from the downstream app
```

## Magic Moment: Deploy the Orchestrator Service with Sidecar Proxy

```sh

cat <<EOF > orchestrator-service.yaml
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: orchestrator-service
  labels:
    cloud.googleapis.com/location: "europe-west1"
spec:
  template:
    spec:
      containers:
        - image: "europe-west1-docker.pkg.dev/$PROJECT_ID/demo-repo/sample-app:latest"
          ports:
            - containerPort: 8080
          env:
            - name: HTTP_PROXY
              value: "http://127.0.0.1:8000"
            - name: TARGET_URL
              value: "$(gcloud run services describe backend-service --region=europe-west1 --project=$PROJECT_ID --format="value(status.url)")/test"
        - image: "europe-west1-docker.pkg.dev/$PROJECT_ID/demo-repo/auth-sidecar:latest"
EOF

gcloud run services replace orchestrator-service.yaml --project $PROJECT_ID
```

try it out: 

```sh
curl "$(gcloud run services describe orchestrator-service --region=europe-west1 --project=$PROJECT_ID --format="value(status.url)")/test" -H "Authorization: Bearer $(gcloud auth print-identity-token)"
```

this should return:

```txt
response from http://<BACKEND-SERVICE-URL>.a.run.app/test: hello from the downstream app
```