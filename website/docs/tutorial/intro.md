---
id: tutorial
title: Overview
sidebar_label: Overview
---
Let's see how TRASA can be used to protect your server and services in the real world scenario.

For the sake of the tutorial, we have picked a hypothetical organization. We will call it Nepsec.
Nepsec is a not-for-profit security organization from Nepal which offers security training and awareness services.

## Nepsec internal infrastructure

Nepsec has the following server and service, which makes up all its internal services.

+ Server 1 - Windows server 2016.
  + Hosted in AWS
  + RDP server listening in port 3389
  
+ Server 2 - Centos 7 server.
  + Hosted in Digital ocean
  + SSH listening in port 22

+ Jenkins Web service installed in GCP Kubernetes


## Nepsec employees

Nepsec has one system administrator and has also hired one 3rd party contractor to manage their AWS and GCP account.
Together, they are the only ones who access their internal infrastructure frequently for maintenance and upgrade.


## Next, we will setup TRASA. 
