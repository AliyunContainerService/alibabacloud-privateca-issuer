# AlibabaCloud Private CA Issuer

English | [简体中文](./README-zh_CN.md)

[AlibabaCloud Private CA Issuer](https://github.com/AliyunContainerService/alibabacloud-privateca-issuer), as an external open-source extension of cert-manager, can help you apply for certificates through [Alibaba Cloud Private CA](https://www.alibabacloud.com/help/en/ssl-certificate/user-guide/private-ca) and store them as TLS secrets in your Kubernetes cluster.

[cert-manager](https://github.com/cert-manager/cert-manager) is an open-source project that helps you create TLS certificates for workloads in your Kubernetes or OpenShift cluster and renews the certificates before they expire.

## Install

1. Ensure that cert-manager is already installed in the cluster

2. Make sure that the credentials used by the **AlibabaCloud Private CA Issuer** has sufficient permissions to access the Alibaba Cloud Private CA service. You can use the following two configuration methods, and we recommend you to use the second **RRSA** method to achieve authorization in the Pod level.

   - Add permissions to the WorkerRole corresponding to the cluster

     - Log in to the Container Service console

     - Select the cluster to enter the cluster details page

     - Navigate to the **Cluster Resources** page in the cluster information. Once there, click on the Worker RAM role with the corresponding name **KubernetesWorkerRole-xxxxxxxxxxxxxxx**. This will automatically take you to the console page associated with the RAM role.

     - Add PCA RAM policy below into the policy bind to the worker role(Only authorize the RAM policy needed for synchronization services, ensuring the principle of minimum permissions.)

       ```
       {
           "Effect": "Allow",
           "Action": [
             "yundun-cert:CreateCustomCertificate",
             "yundun-cert:DescribeCACertificate"
           ],
           "Resource": "*"
       }
       ```

   - Implement Pod dimension authorization through [RRSA method](https://www.alibabacloud.com/help/en/ack/ack-managed-and-ack-dedicated/user-guide/use-rrsa-to-authorize-pods-to-access-different-cloud-services)
     - [Enable RRSA functionality](https://www.alibabacloud.com/help/en/container-service-for-kubernetes/latest/use-rrsa-to-enforce-access-control#section-ywl-59g-j8h)
     - [Use RRSA function](https://www.alibabacloud.com/help/en/container-service-for-kubernetes/latest/use-rrsa-to-enforce-access-control#section-rmr-eeh-878): Create the corresponding RAM role for the specified serviceaccount, set the trust policy for the RAM role, and authorize the RAM role

3. Log in to the Container Service console

   * Select **Marketplace** -> **Marketplace** in the left navigation bar, enter **alibabacloud-privateca-issuer** in the search bar, and select to enter the application page;
   * Select the target cluster, namespace, and release name to be installed;
   * Configure custom parameters on the parameter configuration page. For parameter descriptions, see the **configuration instructions** below;
   * Click the **OK** button to complete the installation.

   If you need to use the RRSA method, please modify the helm values.yaml as follows, refer to [HELM Configuration instructions](#helm-configuration-instructions) for other configuration details.

   ```yaml
   rrsa:
     # Specifies whether using rrsa and enalbe sa token volume projection, default is false
     enable: true
   ```

## Upgrade

1. Log in to the Container Service console;
2. Select the target cluster and click to enter the cluster details page;
3. Select **Applications** -> **Helm** in the navigation bar on the left, find the **Update** button corresponding to **alibabacloud-privateca-issuer**, modify the configuration and click the **OK** button to complete the installation.

## Uninstall

1. Log in to the Container Service console;
2. Select the target cluster and click to enter the cluster details page;
3. Select **Applications** -> **Helm** in the navigation bar on the left, find the **Delete** button corresponding to **alibabacloud-privateca-issuer**, and click the **Delete** button in the operation bar to delete it.

## Helm Configuration Instructions

| parameter                                | introduction                                                 |
| ---------------------------------------- | ------------------------------------------------------------ |
| rbac.create                              | Whether to create and use RBAC resources, the default is true |
| rrsa.enable                              | Whether to enable the RRSA feature, the default is false.    |
| serviceAccount.create                    | Whether to create serviceaccount, the default is true        |
| replicaCount                             | Number of controller copies                                  |
| image.repository                         | Specified alibabacloud-privateca-issuer image warehouse name |
| image.tag                                | Specified alibabacloud-privateca-issuer image tag            |
| image.pullPolicy                         | Image pull strategy, default is Always                       |
| command.region                           | The region where the Kubernetes cluster is located           |
| command.maxConcurrentCertificateRequests | Maximum number of certificate requests per second.           |

## Instructions for use

If you have an available CA certificate in the PCA service and want to issue certificates based on it, please follow the steps below to configure.

1. **Deploy PCAIssuer/PCAClusterIssuer**

   AlibabaCloud Private CA Issuer includes two CRDs (PCAIssuer and PCAClusterIssuer), representing CA certificates available in a PCA service, here is a simple example use RRSA method. (For detailed configuration, please refer to [CRD configuration introduction](#crd-configuration-introduction))

   ```yaml
   apiVersion: 'alibabacloud.com/v1beta'
   kind: PCAClusterIssuer
   metadata:
     name: ca-issuer
   spec:
     parentIdentifier: 1f0169ee-xxxx-xxxx-xxxx-xxxxxxxxxxxx
     ramRoleARN: "acs:ram::xxxxxxxxxxxxxx:role/test-pca"
     ramRoleSessionName: "test-pca"
     oidcProviderARN: acs:ram::xxxxxxxxxxxxxx:oidc-provider/ack-rrsa-{cluster-id}
   ```

2. **Deploy Certificate**, a **Certificate** represents a certificate request, and it needs to reference a specific Issuer to apply for a specific certificate. Here is a simple example(For detailed configuration, please refer to [Certificate](https://cert-manager.io/docs/usage/certificate/))

   ```yaml
   apiVersion: cert-manager.io/v1
   kind: Certificate
   metadata:
     name: pca-certificate
     namespace: default
   spec:
     duration: 1440h
     dnsNames:
       - alibabacloud.com
       - www.example.com
       - fff.com
     emailAddresses:
       - example@aa.com
     ipAddresses:
       - 192.168.0.5
     isCA: false
     usages:
       - server auth
       - digital signature
       - key encipherment
     commonName: pca-certificate-common-name
     secretName: pca-certificate-secret
     privateKey:
       algorithm: ECDSA
       size: 256
     issuerRef:
       name: ca-issuer
       kind: PCAClusterIssuer
       group: alibabacloud.com
   ```

3. Upon completing the above steps, you can see that a TLS secret has been generated in the namespace where the Certificate is located.

## CRD configuration introduction

 **PCAIssuer**/**PCAClusterIssuer** spec

| parameter                | introduction                                               |
| ------------------------ | ---------------------------------------------------------- |
| parentIdentifier         | CA Certificate identifier in the PCA Service.              |
| ramRoleARN               | RAM Role ARN used to access the PCA service.               |
| ramRoleSessionName       | RAM Role session name.                                     |
| oidcProviderARN          | OIDC Provider ARN.                                         |
| remoteRamRoleArn         | Cross-account RAM Role ARN used to access the PCA service. |
| remoteRamRoleSessionName | Cross-account RAM role session name.                       |
