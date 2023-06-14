package config

type WsConfig struct {
	HeartbeatInterval          int      `yaml:"heartbeat_interval"`            // in seconds
	CheckSessionExpireInterval int      `yaml:"check_session_expire_interval"` // in seconds
	PendingSessionCacheTime    int      `yaml:"pending_session_cache_time"`    // in seconds
	MessageCacheTime           int      `yaml:"message_cache_time"`            //
	AllowedOrigins             []string `yaml:"allowed_origins"`
}
