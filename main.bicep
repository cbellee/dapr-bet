param location string
param adminGroupObjectID string
param tags object
param environment string
param aksVersion string = '1.27.3'
param vmSku string = 'Standard_F8s_v2'
param addressPrefix string
param subnets array
param sshPublicKey string
param databases array = [
  {
    name: 'bets'
    partitionKey: '/raceid'
  }
  {
    name: 'punters'
    partitionKey: '/email'
  }
  {
    name: 'results'
    partitionKey: '/raceid'
  }
]

param topics array = [
  'results'
  'payments'
]

module ai 'modules/appInsights.bicep' = {
  name: 'aiDeploy'
  params: {
    location: location
    prefix: environment
  }
}

module wks './modules/wks.bicep' = {
  name: 'wksDeploy'
  params: {
    prefix: environment
    tags: tags
    location: location
    retentionInDays: 30
  }
}

module vnet './modules/vnet.bicep' = {
  name: 'vnetDeploy'
  params: {
    prefix: environment
    tags: tags
    addressPrefix: addressPrefix
    location: location
    subnets: subnets
  }
}

module acr './modules/acr.bicep' = {
  name: 'acrDeploy'
  params: {
    location: location
    prefix: environment
    tags: tags
  }
}

module aks './modules/aks.bicep' = {
  name: 'aksDeploy'
  dependsOn: [
    vnet
    wks
  ]
  params: {
    keyVaultName: keyvault.outputs.name
    enableKeda: true
    location: location
    prefix: environment
    acrName: acr.outputs.name
    logAnalyticsWorkspaceId: wks.outputs.workspaceId
    aksDnsPrefix: environment
    aksAgentOsDiskSizeGB: 60
    aksDnsServiceIP: '10.100.0.10'
    aksDockerBridgeCIDR: '172.17.0.1/16'
    aksEnableRBAC: true
    aksMaxNodeCount: 10
    aksMinNodeCount: 1
    aksNodeCount: 2
    aksNodeVMSize: vmSku
    aksServiceCIDR: '10.100.0.0/16'
    aksSystemSubnetId: vnet.outputs.subnets[0].id
    aksUserSubnetId: vnet.outputs.subnets[1].id
    aksVersion: aksVersion
    enableAutoScaling: true
    maxPods: 110
    networkPlugin: 'azure'
    enablePodSecurityPolicy: false
    tags: tags
    enablePrivateCluster: false
    linuxAdminUserName: 'localadmin'
    sshPublicKey: sshPublicKey
    adminGroupObjectID: adminGroupObjectID
    addOns: {
      azurepolicy: {
        enabled: true
        config: {
          version: 'v2'
        }
      }
      omsagent: {
        enabled: true
        config: {
          logAnalyticsWorkspaceResourceID: wks.outputs.workspaceId
        }
      }
    }
  }
}

module cosmosdb './modules/cosmosDb.bicep' = {
  name: 'cosmosdbDeploy'
  params: {
    location: location
    prefix: environment
    tags: tags
    databases: databases
  }
}

module serviceBus 'modules/sbus.bicep' = {
  name: 'sbusDeploy'
  params: {
    location: location
    prefix: environment
    tags: tags
  }
}

module serviceBusTopicsAndSubs 'modules/sbusTopicsAndSubs.bicep' = {
  name: 'sbusTopicsAndSubsDeploy'
  params: {
    sbusNamespaceName: serviceBus.outputs.name
    topics: topics
  }
}

module keyvault 'modules/kv.bicep' = {
  name: 'kvDeploy'
  params: {
    location: location
    prefix: environment
    tags: tags
  }
}

module cosmosDbConnectionString 'modules/keyVaultSecret.bicep' = {
  name: 'cosmosDbConnectionStringDeploy'
  params: {
    kvName: keyvault.outputs.name
    secret: cosmosdb.outputs.connectionString
    secretName: 'cosmosDbConnectionString'
  }
}

module cosmosDbMasterKey 'modules/keyVaultSecret.bicep' = {
  name: 'cosmosDbMasterKeyDeploy'
  params: {
    kvName: keyvault.outputs.name
    secret: cosmosdb.outputs.masterKey
    secretName: 'cosmosDbMasterKey'
  }
}

module cosmosDbUrl 'modules/keyVaultSecret.bicep' = {
  name: 'cosmosDbUrlDeploy'
  params: {
    kvName: keyvault.outputs.name
    secret: cosmosdb.outputs.url
    secretName: 'cosmosDbUrl'
  }
}

module sbConnectionString 'modules/keyVaultSecret.bicep' = {
  name: 'sbConnectionStringDeploy'
  params: {
    kvName: keyvault.outputs.name
    secret: serviceBus.outputs.connectionString
    secretName: 'sbConnectionString'
  }
}

module aiInstrumentationKey 'modules/keyVaultSecret.bicep' = {
  name: 'aiInstrumentationKeyDeploy'
  params: {
    kvName: keyvault.outputs.name
    secret: ai.outputs.instrumentationKey
    secretName: 'aiInstrumentationKey'
  }
}

output aksClusterName string = aks.outputs.aksClusterName
output aksClusterFqdn string = aks.outputs.aksControlPlaneFQDN
output aksClusterApiServerUri string = aks.outputs.aksApiServerUri
output serviceBusConnectionString string = serviceBus.outputs.connectionString
output acrName string = acr.outputs.name
output midPrincipalId string = aks.outputs.midPrincipalId
output keyVaultName string = keyvault.outputs.name
output cosmosDbKey string = cosmosdb.outputs.masterKey
output cosmosDbUrl string = cosmosdb.outputs.url
