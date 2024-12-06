# Helm Deployment

## Setup

In order to deploy the helm charts, you will need to install the following tools:

1. helm.
2. helmfile.

The helm chart for the custom webservice is located at `./charts/service`.

## Execution

After the tools are installed, you can execute the deployments using:

```sh
export KUBECONFIG="~/.kube/k3s.config"
# at ./helm directory invoke:
helmfile -f helmfile.yml apply --skip-deps --debug
```

This will install the prometheus stack, and the postgres cluster needed.

## Monitoring

### Grafana Dashboard

In order to see Grafana at your host, you can run the following command:

```sh
export KUBECONFIG="~/.kube/k3s.config"
kubectl --context k3s --namespace production port-forward services/prometheus-stack-grafana 8000:80
```

And then go to [Local Grafana](http://localhost:8000) and see the dashboards.

### Project Application Rules, Alerts and Dashboards

The monitroing stack is installed via helm, and the service monitoring rules, alerts and dashboard is located at the chart templates and files. Grafana will automatically download the latest version of service charts via GitHub provisioning link, and you can edit the dashboard in the UI freely after that.

Dashboard is located at `Arvan Service` folder, and `Arvan App Dashboard` dashboard in Grafana.
