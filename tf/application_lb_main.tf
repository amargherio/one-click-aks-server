resource "azurerm_identity" "alb" {
    resource_group_name = azurerm_resource_group.this.name
    name                = module.naming.azurerm_application_load_balancer.name

    depends_on = [
        azurerm_resource_group.this,
        azurerm_kubernetes_cluster.this
    ]
}

resource "azurerm_role_assignment" "alb_reader" {
    principal_id         = azurerm_identity.alb.principal_id
    role_definition_name = "Reader"
    scope                = azurerm_kubernetes_cluster.this.node_resource_group
}

resource "azurerm_federated_identity_credential" "alb" {
    name = module.naming.azurerm_federated_identity_credential.name
    resource_group_name = azurerm_resource_group.this.name
    issuer = azurerm_kubernetes_cluster.this.oidc_issuer_url
    subject = "system:serviceaccount:azure-alb-system:alb-controller"
    parent_id = azurerm_identity.alb.name
    audience = [
        "api://AzureADTokenExchange"
    ]

}

resource helm_release "alb_controller" {
    name = "alb-controller"
    repository = "oci://mcr.microsoft.com/azure/application-lb/charts/alb-controller"
    chart = "alb-controller"
    version = "1.0.0"
    set {
        name = "alb"
        value = azurerm_identity.alb.client_id
    }
}

resource "azurerm_application_load_balancer" "this" {
    name = module.naming.azurerm_application_load_balancer.name

    location = azurerm_resource_group.this.location
    resource_group_name = azurerm_resource_group.this.name
}