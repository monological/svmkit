package watchtower

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/abklabs/svmkit/pkg/runner"
	"github.com/abklabs/svmkit/pkg/solana"
)

type InstallCommand struct {
	Watchtower
}

func (cmd *InstallCommand) Env() *runner.EnvBuilder {
	watchtowerEnv := runner.NewEnvBuilder()

	if cmd.Notifications.Slack != nil {
		watchtowerEnv.Set("SLACK_WEBHOOK", cmd.Notifications.Slack.WebhookURL)
	}

	if cmd.Notifications.Discord != nil {
		watchtowerEnv.Set("DISCORD_WEBHOOK", cmd.Notifications.Discord.WebhookURL)
	}

	if cmd.Notifications.Telegram != nil {
		watchtowerEnv.Set("TELEGRAM_BOT_TOKEN", cmd.Notifications.Telegram.BotToken)
		watchtowerEnv.Set("TELEGRAM_CHAT_ID", cmd.Notifications.Telegram.ChatID)
	}

	if cmd.Notifications.PagerDuty != nil {
		watchtowerEnv.Set("PAGERDUTY_INTEGRATION_KEY", cmd.Notifications.PagerDuty.IntegrationKey)
	}

	if cmd.Notifications.Twilio != nil {
		watchtowerEnv.Set("TWILIO_CONFIG", cmd.Notifications.Twilio.String())
	}

	b := runner.NewEnvBuilder()

	b.SetMap(map[string]string{
		"WATCHTOWER_FLAGS": strings.Join(cmd.ToArgs(), " "),
		"WATCHTOWER_ENV":   watchtowerEnv.String(),
	})

	return b

}

func (cmd *InstallCommand) Check() error {
	return nil
}

func (cmd *InstallCommand) AddToPayload(p *runner.Payload) error {
	err := p.AddTemplate("steps.sh", installScriptTmpl, cmd)

	if err != nil {
		return err
	}

	return nil
}

type Watchtower struct {
	Environment   solana.Environment `pulumi:"environment"`
	Flags         Flags              `pulumi:"flags"`
	Notifications NotificationConfig `pulumi:"notifications"`
}

func (w *Watchtower) ToArgs() []string {
	return w.Flags.ToArgs(w.Environment.RPCURL)
}

func (w *Watchtower) Install() runner.Command {
	return &InstallCommand{
		Watchtower: *w,
	}
}

type Flags struct {
	IntervalSeconds     *int     `pulumi:"intervalSeconds,optional"`
	RPCTimeoutSeconds   *int     `pulumi:"rpcTimeoutSeconds,optional"`
	UnhealthyThreshold  *int     `pulumi:"unhealthyThreshold,optional"`
	ValidatorIdentities []string `pulumi:"validatorIdentities"`
	NameSuffix          *string  `pulumi:"nameSuffix,optional"`

	MonitorActiveStake          *bool `pulumi:"monitorActiveStake,optional"`
	ActiveStakeAlertThreshold   *int  `pulumi:"activeStakeAlertThreshold,optional"`
	MinValidatorIdentityBalance *int  `pulumi:"minimumValidatorIdentityBalance,optional"`
	IgnoreHTTPBadGateway        *bool `pulumi:"ignoreHttpBadGateway,optional"`
}

func (f *Flags) ToArgs(rpcURL *string) []string {
	b := runner.FlagBuilder{}

	for _, identity := range f.ValidatorIdentities {
		b.Append("--validator-identity", identity)
	}

	if rpcURL != nil {
		b.AppendP("url", rpcURL)
	}

	b.AppendIntP("interval", f.IntervalSeconds)
	b.AppendIntP("rpc-timeout", f.RPCTimeoutSeconds)
	b.AppendIntP("unhealthy-threshold", f.UnhealthyThreshold)

	b.AppendP("name-suffix", f.NameSuffix)
	b.AppendIntP("active-stake-alert-threshold", f.ActiveStakeAlertThreshold)
	b.AppendIntP("minimum-validator-identity-balance", f.MinValidatorIdentityBalance)

	b.AppendBoolP("--monitor-active-stake", f.MonitorActiveStake)
	b.AppendBoolP("--ignore-http-bad-gateway", f.IgnoreHTTPBadGateway)

	return b.ToArgs()
}

type NotificationConfig struct {
	Slack     *SlackConfig     `pulumi:"slack,optional"`
	Discord   *DiscordConfig   `pulumi:"discord,optional"`
	Telegram  *TelegramConfig  `pulumi:"telegram,optional"`
	PagerDuty *PagerDutyConfig `pulumi:"pagerDuty,optional"`
	Twilio    *TwilioConfig    `pulumi:"twilio,optional"`
}

type SlackConfig struct {
	WebhookURL string `pulumi:"webhookUrl"`
}

type DiscordConfig struct {
	WebhookURL string `pulumi:"webhookUrl"`
}

type TelegramConfig struct {
	BotToken string `pulumi:"botToken"`
	ChatID   string `pulumi:"chatId"`
}

type PagerDutyConfig struct {
	IntegrationKey string `pulumi:"integrationKey"`
}

type TwilioConfig struct {
	AccountSID string `pulumi:"accountSid"`
	AuthToken  string `pulumi:"authToken"`
	ToNumber   string `pulumi:"toNumber"`
	FromNumber string `pulumi:"fromNumber"`
}

func (t *TwilioConfig) String() string {

	configParts := []string{
		fmt.Sprintf("ACCOUNT=%s", t.AccountSID),
		fmt.Sprintf("TOKEN=%s", t.AuthToken),
		fmt.Sprintf("TO=%s", t.ToNumber),
		fmt.Sprintf("FROM=%s", t.FromNumber),
	}

	return strings.Join(configParts, ",")
}
