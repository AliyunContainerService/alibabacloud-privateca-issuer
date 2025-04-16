# AlibabaCloud Private CA Issuer

[AlibabaCloud Private CA Issuer](https://github.com/AliyunContainerService/alibabacloud-privateca-issuer) 是 cert-manager 的一个外部开源扩展，能够帮助您通过 [阿里云证书服务](https://help.aliyun.com/zh/ssl-certificate/user-guide/private-ca/) 申请私有证书，并将其存储在Kubernetes 集群的 TLS Secret 中。

[cert-manager](https://github.com/cert-manager/cert-manager) 是一个开源项目，可以为 Kubernetes 集群中的工作负载创建 TLS 证书，并进行证书全生命周期的管理。

## Install

1. 通过 **[install](https://cert-manager.io/docs/installation/)** 安装 cert-manager

2. 为确保 **AlibabaCloud Private CA Issuer** 使用的凭据具有足够的权限来访问阿里云证书服务，可以使用如下两种配置方式，推荐使用RRSA方式，实现Pod维度的授权

    - 在集群对应的 WorkerRole 中添加权限

        - 登录容器服务控制台

        - 选择对应集群进入到集群详情页

        - 在集群信息中选择**集群资源**页，点击Worker RAM角色中对应的命名为**KubernetesWorkerRole-xxxxxxxxxxxxxxx** 的角色名称，会自动导航到RAM角色对应的控制台页面

        - 点击添加权限按钮，创建自定义权限策略，策略内容如下（仅授权组件需要的RAM策略即可，保证最小权限原则）：

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

    - 通过 [RRSA方式](https://help.aliyun.com/document_detail/356611.html) 实现Pod维度的授权
        - [启用RRSA功能](https://help.aliyun.com/document_detail/356611.html#section-ywl-59g-j8h)
        - [使用RRSA功能](https://help.aliyun.com/document_detail/356611.html#section-rmr-eeh-878) ：为指定的 serviceaccount 创建对应的 RAM 角色，为 RAM 角色设置信任策略，并为 RAM 角色授权

3. 登录到容器服务控制台

    * 在左侧导航栏选择**市场** -> **应用市场**，在搜索栏中输入**alibabacloud-privateca-issuer**，选择进入到应用页面；
    * 选择需要安装的目标集群和命名空间、发布名称；
    * 在参数配置页面进行自定义参数配置。 对于详细的参数配置请参考[Helm 详细配置](#helm-详细配置)
    * 点击**确定**按钮完成安装。


## 升级

1. 登录到容器服务控制台；
2. 选择目标集群点击进入到集群详情页面；
3. 在左侧的导航栏选择应用-> Helm，找到 **alibabacloud-privateca-issuer** 对应的**更新**，修改配置后点击**确定**按钮完成安装。

## 卸载

1. 登录到容器服务控制台；
2. 选择目标集群点击进入到集群详情页面；
3. 在左侧的导航栏选择应用-> Helm，找到 ack-secret-manager 对应的发布，点击操作拦中的删除按钮进行删除。

## Helm 详细配置

| 参数                                     | 说明                                             |
| ---------------------------------------- | ------------------------------------------------ |
| rbac.create                              | 是否创建并使用 RBAC 资源，默认为true             |
| rrsa.enable                              | 是否启用 RRSA 特性，默认为false                  |
| serviceAccount.create                    | 是否创建 serviceaccount，默认为true              |
| replicaCount                             | 控制器副本个数                                   |
| image.repository                         | 指定的 alibabcloud-privateca-issuer 镜像仓库名称 |
| image.tag                                | 指定的 alibabcloud-privateca-issuer 镜像tag      |
| image.pullPolicy                         | 镜像拉取策略，默认为 IfNotPresent                |
| command.region                           | Kubernetes集群所在region                         |
| command.maxConcurrentCertificateRequests | 每秒最大处理的证书请求数量                       |

## 使用说明

如果您在 PCA 服务中有可用的 CA 证书，并希望基于其签发证书，请按照以下步骤进行配置。

1. **部署 PCAIssuer/PCAClusterIssuer**

   AlibabaCloud Private CA Issuer 包含两种CRD (PCAIssuer and PCAClusterIssuer), 每个实例化后的CR代表一个PCA服务中可用的CA。 以下是简单样例 (对于 PCAIssuer 以及 PCAClusterIssuer 的 详细配置，参考 [CRD 配置说明](#crd-配置说明))

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

2. **部署 Certificate**, 一个 **Certificate** 代表一个证书请求, 并且必须关联一个Issuer实例用于请求特定的证书. 以下是一个简单的样例(对于Certificate更详细的配置方式，请参考 [Certificate](https://cert-manager.io/docs/usage/certificate/))

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

3. 在执行完上述步骤后，可以看到Certificate所在的命名空间生成了对应的 TLS Secret

## CRD 配置说明

**PCAIssuer**/**PCAClusterIssuer** spec

| 参数                     | 说明                           |
| ------------------------ | ------------------------------ |
| parentIdentifier         | CA证书在PCA服务中的证书识别码  |
| ramRoleARN               | 用于访问PCA服务的RAM角色       |
| ramRoleSessionName       | RAM角色会话名                  |
| oidcProviderARN          | OIDC 提供商 ARN                |
| remoteRamRoleArn         | 用于访问PCA服务的跨账号RAM角色 |
| remoteRamRoleSessionName | 跨账号 RAM 角色 session name   |
