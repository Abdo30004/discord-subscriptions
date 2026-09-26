package k8s

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/discord-subscriptions/deploy-svc/internal/core/domain"
	"github.com/discord-subscriptions/deploy-svc/internal/core/ports"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	clientset  *kubernetes.Clientset
	logger     *slog.Logger
	isMockMode bool
}

// NewClient initializes a Kubernetes API client using in-cluster configuration or local kubeconfig.
func NewClient(kubeconfigPath string, logger *slog.Logger) (*Client, error) {
	if logger == nil {
		logger = slog.Default()
	}

	// Try in-cluster config first
	config, err := rest.InClusterConfig()
	if err != nil {
		// Fallback to kubeconfig file
		if kubeconfigPath != "" {
			config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		}
	}

	if err != nil {
		logger.Warn("Kubernetes cluster not detected, running in simulation mode for local development", slog.String("error", err.Error()))
		return &Client{
			clientset:  nil,
			logger:     logger,
			isMockMode: true,
		}, nil
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed creating kubernetes clientset: %w", err)
	}

	return &Client{
		clientset:  clientset,
		logger:     logger,
		isMockMode: false,
	}, nil
}

// DeployBot creates or updates a Kubernetes Deployment for a dedicated bot instance.
func (c *Client) DeployBot(ctx context.Context, params ports.BotDeployParams) error {
	if c.isMockMode {
		c.logger.Info("[DEV SIMULATION] Deployed bot container to K8s",
			slog.String("namespace", params.Namespace),
			slog.String("deployment", params.DeploymentName),
			slog.String("image", params.Image),
		)
		return nil
	}

	// 1. Ensure Namespace exists
	if err := c.ensureNamespace(ctx, params.Namespace); err != nil {
		return fmt.Errorf("failed ensuring namespace %s: %w", params.Namespace, err)
	}

	// 2. Build Deployment specification
	replicas := int32(1)
	labels := map[string]string{
		"app.kubernetes.io/name":       "discord-bot",
		"app.kubernetes.io/instance":   params.DeploymentName,
		"discord.subscriptions/guild":  params.GuildID,
		"discord.subscriptions/bot-id": params.BotID,
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      params.DeploymentName,
			Namespace: params.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "bot",
							Image: params.Image,
							Env: []corev1.EnvVar{
								{Name: "DISCORD_TOKEN", Value: params.BotToken},
								{Name: "BOT_ID", Value: params.BotID},
								{Name: "GUILD_ID", Value: params.GuildID},
								{Name: "PORT", Value: "8080"},
							},
							Ports: []corev1.ContainerPort{
								{Name: "health", ContainerPort: 8080},
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/health",
										Port: intstr.FromInt32(8080),
									},
								},
								InitialDelaySeconds: 15,
								PeriodSeconds:       10,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/health",
										Port: intstr.FromInt32(8080),
									},
								},
								InitialDelaySeconds: 5,
								PeriodSeconds:       5,
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("250m"),
									corev1.ResourceMemory: resource.MustParse("256Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("50m"),
									corev1.ResourceMemory: resource.MustParse("64Mi"),
								},
							},
						},
					},
					RestartPolicy: corev1.RestartPolicyAlways,
				},
			},
		},
	}

	deploymentsClient := c.clientset.AppsV1().Deployments(params.Namespace)
	_, err := deploymentsClient.Get(ctx, params.DeploymentName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		_, err = deploymentsClient.Create(ctx, deployment, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed creating k8s deployment: %w", err)
		}
		c.logger.Info("created k8s deployment", slog.String("name", params.DeploymentName))
		return nil
	} else if err != nil {
		return fmt.Errorf("error checking existing deployment: %w", err)
	}

	// Update existing deployment
	_, err = deploymentsClient.Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed updating k8s deployment: %w", err)
	}
	c.logger.Info("updated k8s deployment", slog.String("name", params.DeploymentName))

	return nil
}

// StopBot scales the deployment to 0 replicas.
func (c *Client) StopBot(ctx context.Context, namespace, deploymentName string) error {
	if c.isMockMode {
		c.logger.Info("[DEV SIMULATION] Stopped bot deployment", slog.String("name", deploymentName))
		return nil
	}

	scale, err := c.clientset.AppsV1().Deployments(namespace).GetScale(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed fetching deployment scale: %w", err)
	}

	scale.Spec.Replicas = 0
	_, err = c.clientset.AppsV1().Deployments(namespace).UpdateScale(ctx, deploymentName, scale, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed setting replicas to 0: %w", err)
	}

	return nil
}

// RestartBot triggers a rolling restart by updating an annotation on the Pod template.
func (c *Client) RestartBot(ctx context.Context, namespace, deploymentName string) error {
	if c.isMockMode {
		c.logger.Info("[DEV SIMULATION] Restarted bot deployment", slog.String("name", deploymentName))
		return nil
	}

	deployment, err := c.clientset.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed fetching deployment for restart: %w", err)
	}

	if deployment.Spec.Template.Annotations == nil {
		deployment.Spec.Template.Annotations = make(map[string]string)
	}
	deployment.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)

	_, err = c.clientset.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed triggering rollout restart: %w", err)
	}

	return nil
}

// GetBotStatus reads the active replicas and conditions from Kubernetes.
func (c *Client) GetBotStatus(ctx context.Context, namespace, deploymentName string) (domain.DeploymentStatus, error) {
	if c.isMockMode {
		return domain.StatusRunning, nil
	}

	dep, err := c.clientset.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return domain.StatusStopped, nil
		}
		return domain.StatusFailed, err
	}

	if dep.Status.ReadyReplicas > 0 {
		return domain.StatusRunning, nil
	}
	if dep.Status.Replicas == 0 {
		return domain.StatusStopped, nil
	}

	return domain.StatusDeploying, nil
}

func (c *Client) ensureNamespace(ctx context.Context, namespace string) error {
	_, err := c.clientset.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace,
			},
		}
		_, err = c.clientset.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
		return err
	}
	return err
}
