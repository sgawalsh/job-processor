$timeout = 120

Write-Host "Bootstrapping Kubernetes Cluster..."

############################################
# 1. Namespaces
############################################
Write-Host "Creating namespaces..."
kubectl apply -R -f .\namespaces\

############################################
# 2. Install Monitoring (Helm)
############################################
Write-Host "Installing kube-prometheus-stack..."

helm repo add prometheus-community https://prometheus-community.github.io/helm-charts 2>$null
helm repo update

helm upgrade --install monitoring prometheus-community/kube-prometheus-stack `
  --namespace monitoring `
  --create-namespace `
  -f .\monitoring\values.yaml

############################################
# 3. Wait for Monitoring Pods
############################################
Write-Host "Waiting for monitoring pods..."
kubectl wait --for=condition=Ready pods --all -n monitoring --timeout=${timeout}s

############################################
# 4. Deploy App
############################################
Write-Host "Deploying application..."
kubectl apply -R -f .\app\

kubectl wait --for=condition=Ready pods --all -n app --timeout=${timeout}s

############################################
# 5. Install KEDA and deploy ScaledObject
############################################
Write-Host "Installing KEDA..."

helm repo add kedacore https://kedacore.github.io/charts 2>$null
helm repo update

helm upgrade --install keda kedacore/keda `
  --namespace keda `
  --create-namespace

Write-Host "Waiting for KEDA operator..."
kubectl wait --for=condition=Available deployment/keda-operator -n keda --timeout=${timeout}s

Write-Host "Deploying KEDA ScaledObject..."
kubectl apply -R -f .\keda\

############################################
# 6. Apply Monitoring Resources
############################################
Write-Host "Applying ServiceMonitors and dashboards..."
kubectl apply -R -f .\monitoring\monitors\
kubectl apply -R -f .\monitoring\dashboards\

############################################
# 7. Install Ingress Controller
############################################
Write-Host "Installing ingress controller..."
kubectl apply -f .\ingress\controller.yaml

Write-Host "Waiting for ingress controller deployment..."

kubectl wait --namespace ingress-nginx `
  --for=condition=available deployment/ingress-nginx-controller `
  --timeout=120s

if ($LASTEXITCODE -ne 0) {
    Write-Error "Ingress controller failed to become ready."
    exit 1
}

Write-Host "Ingress controller is ready."

Write-Host "Waiting for admission webhook service endpoints..."

$timeoutSeconds = 60
$interval = 3
$elapsed = 0

do {
    $endpoints = kubectl get endpoints ingress-nginx-controller-admission `
        -n ingress-nginx `
        -o jsonpath='{.subsets[*].addresses[*].ip}'

    if ($endpoints) { break }

    Start-Sleep -Seconds $interval
    $elapsed += $interval

} while ($elapsed -lt $timeoutSeconds)

if (-not $endpoints) {
    Write-Error "Admission webhook endpoints not ready."
    exit 1
}

Write-Host "Admission webhook is ready."



############################################
# 8. Apply Ingress Resource
############################################
Write-Host "Applying ingress resources..."

$interval = 5
$elapsed = 0
$success = $false

do {
    kubectl apply -f .\ingress\api-ingress.yaml 2>$null
    kubectl apply -f .\ingress\frontend-ingress.yaml 2>$null

    if ($LASTEXITCODE -eq 0) {
        $success = $true
        break
    }

    Write-Host "Webhook not ready yet, retrying..."
    Start-Sleep -Seconds $interval
    $elapsed += $interval

} while ($elapsed -lt $timeout)

if (-not $success) {
    Write-Error "Failed to apply ingress resources after waiting."
    exit 1
}

Write-Host "Ingress resources applied successfully."

Write-Host "Cluster bootstrap complete"
