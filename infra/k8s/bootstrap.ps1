$timeout = 60

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
# 5. Apply Monitoring Resources
############################################
Write-Host "Applying ServiceMonitors and dashboards..."
kubectl apply -R -f .\monitoring\monitors\
kubectl apply -R -f .\monitoring\dashboards\

############################################
# 6. Install Ingress Controller
############################################
Write-Host "Installing ingress controller..."
kubectl apply -R -f .\ingress\

$interval = 5
$elapsed = 0

Write-Host "Waiting for ingress controller pod and webhook service..."

do {
    $podReady = kubectl get pods -n ingress-nginx -l app.kubernetes.io/name=ingress-nginx -o json | ConvertFrom-Json | `
        ForEach-Object { $_.status.conditions | Where-Object { $_.type -eq "Ready" -and $_.status -eq "True" } } | Measure-Object | Select-Object -ExpandProperty Count

    $podCount = (kubectl get pods -n ingress-nginx -l app.kubernetes.io/name=ingress-nginx | Select-Object -Skip 1 | Measure-Object).Count

    if ($podReady -eq $podCount -and $podCount -gt 0) { break }

    Start-Sleep -Seconds $interval
    $elapsed += $interval
} while ($elapsed -lt $timeout)

Write-Host "Ingress controller pods are ready."
kubectl apply -f .\ingress\ingress.yaml

############################################
# 7. Apply Ingress Resource
############################################
Write-Host "Applying ingress resource..."
kubectl apply -f .\ingress\ingress.yaml

Write-Host "Cluster bootstrap complete"
