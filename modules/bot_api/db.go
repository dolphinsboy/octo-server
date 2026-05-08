package bot_api

import (
	"github.com/Mininglamp-OSS/octo-lib/config"
	"github.com/Mininglamp-OSS/octo-lib/pkg/db"
	"github.com/gocraft/dbr/v2"
)

type botAPIDB struct {
	session *dbr.Session
	ctx     *config.Context
}

func newBotAPIDB(ctx *config.Context) *botAPIDB {
	return &botAPIDB{
		ctx:     ctx,
		session: ctx.DB(),
	}
}

// ==================== Robot Model (User Bot) ====================

type robotModel struct {
	AppID         string
	RobotID       string
	Username      string
	InlineOn      int
	Placeholder   string
	Token         string
	Version       int64
	Status        int
	CreatorUID    string
	Description   string
	BotToken      string
	IMTokenCache  string
	BotCommands   string
	AutoApprove   int
	AccessMode    int
	AgentPlatform string
	AgentVersion  string
	PluginVersion string
	db.BaseModel
}

// queryRobotByBotToken queries robot by bot token.
func (d *botAPIDB) queryRobotByBotToken(botToken string) (*robotModel, error) {
	if botToken == "" {
		return nil, nil
	}
	var m *robotModel
	_, err := d.session.Select("*").From("robot").Where("bot_token=? and bot_token!='' and status=1", botToken).Load(&m)
	return m, err
}

// queryRobotByRobotID queries robot by robot ID.
func (d *botAPIDB) queryRobotByRobotID(robotID string) (*robotModel, error) {
	var m *robotModel
	_, err := d.session.Select("*").From("robot").Where("robot_id=?", robotID).Load(&m)
	return m, err
}

// updateRobotIMTokenCache updates the IM token cache for a robot.
func (d *botAPIDB) updateRobotIMTokenCache(robotID string, imToken string) error {
	_, err := d.session.Update("robot").SetMap(map[string]interface{}{
		"im_token_cache": imToken,
	}).Where("robot_id=?", robotID).Exec()
	return err
}

// updateRobotAgentInfo updates agent runtime info for a robot.
func (d *botAPIDB) updateRobotAgentInfo(robotID, agentPlatform, agentVersion, pluginVersion string) error {
	_, err := d.session.Update("robot").SetMap(map[string]interface{}{
		"agent_platform": agentPlatform,
		"agent_version":  agentVersion,
		"plugin_version": pluginVersion,
	}).Where("robot_id=?", robotID).Exec()
	return err
}

// updateBotCommands updates bot commands JSON.
func (d *botAPIDB) updateBotCommands(robotID string, botCommands string) error {
	_, err := d.session.Update("robot").SetMap(map[string]interface{}{
		"bot_commands": botCommands,
	}).Where("robot_id=?", robotID).Exec()
	return err
}

// queryAllActiveRobots queries all active robots with non-empty bot_token.
func (d *botAPIDB) queryAllActiveRobots() ([]*robotModel, error) {
	var models []*robotModel
	_, err := d.session.Select("*").From("robot").Where("status=1 AND bot_token != ''").Load(&models)
	return models, err
}

// querySpaceIDByRobotID returns the active Space ID for the given bot.
func (d *botAPIDB) querySpaceIDByRobotID(robotID string) (string, error) {
	var spaceID string
	err := d.session.SelectBySql(
		"SELECT sm.space_id FROM space_member sm INNER JOIN space s ON s.space_id = sm.space_id WHERE sm.uid=? AND sm.status=1 AND s.status=1",
		robotID,
	).LoadOne(&spaceID)
	return spaceID, err
}

// ==================== App Bot Model ====================

type appBotModel struct {
	ID          string `db:"id"`
	UID         string `db:"uid"`
	DisplayName string `db:"display_name"`
	Description string `db:"description"`
	Avatar      string `db:"avatar"`
	Scope       string `db:"scope"`
	SpaceID     string `db:"space_id"`
	Status      int    `db:"status"`
	Token       string `db:"token"`
	CreatedBy   string `db:"created_by"`
	db.BaseModel
}

// queryAppBotByToken queries app_bot by token.
func (d *botAPIDB) queryAppBotByToken(token string) (*appBotModel, error) {
	if token == "" {
		return nil, nil
	}
	var m *appBotModel
	_, err := d.session.Select("*").From("app_bot").Where("token=?", token).Load(&m)
	return m, err
}

// queryAppBotByUID queries app_bot by UID.
func (d *botAPIDB) queryAppBotByUID(uid string) (*appBotModel, error) {
	var m *appBotModel
	_, err := d.session.Select("*").From("app_bot").Where("uid=?", uid).Load(&m)
	return m, err
}
