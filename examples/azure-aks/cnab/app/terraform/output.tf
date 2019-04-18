// TODO: designate appropriate values below as sensitive? (Perhaps optionally via bundle/mixin logic)
//
// See https://www.terraform.io/docs/configuration/outputs.html#sensitive-suppressing-values-in-cli-output
// 
// Otherwise, values will be printed out to console after each apply

output "client_key" {
    value = "${azurerm_kubernetes_cluster.k8s.kube_config.0.client_key}"
}

output "client_certificate" {
    value = "${azurerm_kubernetes_cluster.k8s.kube_config.0.client_certificate}"
}

output "cluster_ca_certificate" {
    value = "${azurerm_kubernetes_cluster.k8s.kube_config.0.cluster_ca_certificate}"
}

output "cluster_username" {
    value = "${azurerm_kubernetes_cluster.k8s.kube_config.0.username}"
}

output "cluster_password" {
    value = "${azurerm_kubernetes_cluster.k8s.kube_config.0.password}"
}

output "kube_config" {
    value = "${azurerm_kubernetes_cluster.k8s.kube_config_raw}"
}

output "host" {
    value = "${azurerm_kubernetes_cluster.k8s.kube_config.0.host}"
}