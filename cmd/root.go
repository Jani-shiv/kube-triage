package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"kube-triage/internal/analyzer"
	"kube-triage/internal/collector"
	"kube-triage/internal/renderer"
)

var (
	cfgFile       string
	kubeconfig    string
	namespace     string
	contextName   string
	dryRun        bool
	llmProvider   string
	llmModel      string
	llmEndpoint   string
	previousLogs  bool
	tailLines     int64
	outputFormat  string
)

var rootCmd = &cobra.Command{
	Use:   "kube-triage",
	Short: "AI-powered Kubernetes diagnostic engine",
	Long: `kube-triage analyzes failing Kubernetes resources, pinpoints root causes,
and provides precise, production-safe remediation commands.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initConfig()
	},
}

var analyzeCmd = &cobra.Command{
	Use:   "analyze <resource-type> <resource-name>",
	Short: "Analyze a failing Kubernetes resource",
	Long:  `Analyze a pod, deployment, statefulset, or other resource to diagnose failures.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runAnalyze,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("kube-triage v1.0.0")
		fmt.Println("AI Kubernetes Diagnostic Engine")
	},
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: $HOME/.kube-triage.yaml)")
	rootCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", "", "path to kubeconfig file")
	rootCmd.PersistentFlags().StringVar(&namespace, "namespace", "default", "Kubernetes namespace")
	rootCmd.PersistentFlags().StringVar(&contextName, "context", "", "Kubernetes context to use")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "show remediation commands without executing")
	rootCmd.PersistentFlags().StringVar(&llmProvider, "llm-provider", "auto", "LLM provider: auto, nvidia, openai, anthropic, ollama, none")
	rootCmd.PersistentFlags().StringVar(&llmModel, "llm-model", "", "LLM model name (provider-specific)")
	rootCmd.PersistentFlags().StringVar(&llmEndpoint, "llm-endpoint", "", "Custom LLM endpoint URL")
	rootCmd.PersistentFlags().BoolVar(&previousLogs, "previous", false, "include previous container logs (for crashed containers)")
	rootCmd.PersistentFlags().Int64Var(&tailLines, "tail", 100, "number of log lines to fetch")
	rootCmd.PersistentFlags().StringVar(&outputFormat, "output", "terminal", "output format: terminal, json, yaml")

	analyzeCmd.Flags().StringVar(&namespace, "namespace", "default", "Kubernetes namespace")
	analyzeCmd.Flags().BoolVar(&previousLogs, "previous", false, "include previous container logs")
	analyzeCmd.Flags().Int64Var(&tailLines, "tail", 100, "number of log lines to fetch")

	rootCmd.AddCommand(analyzeCmd)
	rootCmd.AddCommand(versionCmd)

	viper.BindPFlag("kubeconfig", rootCmd.PersistentFlags().Lookup("kubeconfig"))
	viper.BindPFlag("namespace", rootCmd.PersistentFlags().Lookup("namespace"))
	viper.BindPFlag("context", rootCmd.PersistentFlags().Lookup("context"))
	viper.BindPFlag("dry-run", rootCmd.PersistentFlags().Lookup("dry-run"))
	viper.BindPFlag("llm-provider", rootCmd.PersistentFlags().Lookup("llm-provider"))
	viper.BindPFlag("llm-model", rootCmd.PersistentFlags().Lookup("llm-model"))
	viper.BindPFlag("llm-endpoint", rootCmd.PersistentFlags().Lookup("llm-endpoint"))
	viper.BindPFlag("previous", rootCmd.PersistentFlags().Lookup("previous"))
	viper.BindPFlag("tail", rootCmd.PersistentFlags().Lookup("tail"))
	viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
}

func initConfig() error {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		viper.AddConfigPath(home)
		viper.SetConfigName(".kube-triage")
		viper.SetConfigType("yaml")
	}

	viper.SetEnvPrefix("KUBE_TRIAGE")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}
	return nil
}

func runAnalyze(cmd *cobra.Command, args []string) error {
	resourceType := args[0]
	resourceName := args[1]

	ns := viper.GetString("namespace")
	ctx := viper.GetString("context")
	kubeCfg := viper.GetString("kubeconfig")
	prev := viper.GetBool("previous")
	tail := viper.GetInt64("tail")
	dryRun := viper.GetBool("dry-run")
	provider := viper.GetString("llm-provider")
	model := viper.GetString("llm-model")
	endpoint := viper.GetString("llm-endpoint")
	output := viper.GetString("output")

	client, err := collector.NewClient(kubeCfg, ctx)
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	c := collector.NewCollector(client, ns, prev, tail)

	diagData, err := c.Collect(cmd.Context(), resourceType, resourceName)
	if err != nil {
		return fmt.Errorf("collection failed: %w", err)
	}

	engine := analyzer.NewEngine(provider, model, endpoint)
	result, err := engine.Analyze(cmd.Context(), diagData)
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	printer := renderer.NewPrinter(output)
	return printer.Render(result, dryRun)
}

func Execute() error {
	return rootCmd.Execute()
}