# Which EC2 instances Auto Mode may launch for our pods.
#
# Auto Mode picks instances with Karpenter, from NodePools. Its built-in
# general-purpose pool only allows on-demand c/m/r instances, so the cheapest
# it can pick is a .large (it chose a c5a.large, ~$60/month). This pool allows
# small burstable t3/t3a instances instead, spot first (like the GKE stack's
# spot node), falling back to on-demand when spot isn't available. Karpenter
# picks the cheapest instance that fits the pending pods.
#
# general-purpose is turned off in eks.tf, so this is the only pool app pods
# can run on. The built-in system pool stays on: it keeps Auto Mode's default
# NodeClass (subnets, security groups, node role) in place, and it's tainted
# for Auto Mode's own components, so app pods never land on it.
resource "kubectl_manifest" "nodepool_small" {
  count = local.k8s_count

  yaml_body = yamlencode({
    apiVersion = "karpenter.sh/v1"
    kind       = "NodePool"
    metadata = {
      name = "small"
    }
    spec = {
      template = {
        spec = {
          nodeClassRef = {
            group = "eks.amazonaws.com"
            kind  = "NodeClass"
            name  = "default"
          }
          requirements = [
            {
              key      = "karpenter.sh/capacity-type"
              operator = "In"
              values   = ["spot", "on-demand"]
            },
            {
              # t3a (AMD) is a little cheaper where it's sold; eu-north-1
              # only has t3. A spot t3.small there is ~$0.003/hour.
              key      = "eks.amazonaws.com/instance-family"
              operator = "In"
              values   = ["t3", "t3a"]
            },
            {
              key      = "eks.amazonaws.com/instance-size"
              operator = "In"
              values   = ["small", "medium"]
            },
            {
              # The images are built for amd64 (docker/hello2/Dockerfile).
              key      = "kubernetes.io/arch"
              operator = "In"
              values   = ["amd64"]
            },
          ]
          # Replace nodes every two weeks, as the built-in pools do, so they
          # pick up new AMIs.
          expireAfter = "336h"
        }
      }
      # A ceiling on the whole pool: at most 4 vCPUs across all its nodes
      # (two t3a.medium), whatever gets scheduled.
      limits = {
        cpu = "4"
      }
      disruption = {
        # Remove empty nodes and move pods onto fewer/cheaper ones.
        consolidationPolicy = "WhenEmptyOrUnderutilized"
        consolidateAfter    = "30s"
      }
    }
  })

  depends_on = [time_sleep.access_propagation]
}
