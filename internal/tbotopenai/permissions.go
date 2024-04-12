package tbotopenai

const anyUser = "*"

func (t *TBotOpenAI) setUserRoles(roles *RolesSettings) {
	admins := make(map[string]struct{}, len(roles.Admins))
	for _, username := range roles.Admins {
		admins[username] = struct{}{}
	}
	users := make(map[string]struct{}, len(roles.Users))
	for _, username := range roles.Users {
		users[username] = struct{}{}
	}
	t.userRoles.Store(roleAdmin, admins)
	t.userRoles.Store(roleUser, users)
}

func (t *TBotOpenAI) setPermissions(permissions *PermissionSettings) {
	adminCommands := make(map[string]struct{}, len(permissions.AdminCommands))
	for _, command := range permissions.AdminCommands {
		adminCommands[command] = struct{}{}
	}
	userCommands := make(map[string]struct{}, len(permissions.UserCommands))
	for _, command := range permissions.UserCommands {
		userCommands[command] = struct{}{}
	}
	t.permissions.Store(roleAdmin, adminCommands)
	t.permissions.Store(roleUser, userCommands)
}

func (t *TBotOpenAI) checkPermissions(command, username string) bool {
	curRole := t.getRole(username)
	if curRole == "" {
		return false
	}
	val, ok := t.permissions.Load(curRole)
	if !ok {
		return false
	}
	permissions, ok := val.(map[string]struct{})
	if !ok {
		return false
	}
	if _, ok = permissions[anyUser]; ok {
		return true
	}
	if _, ok = permissions[command]; ok {
		return true
	}
	return false
}

func (t *TBotOpenAI) getRole(username string) string {
	var curRole string
	for _, role := range []string{roleAdmin, roleUser} {
		val, ok := t.userRoles.Load(role)
		if !ok {
			continue
		}
		roles, ok := val.(map[string]struct{})
		if !ok {
			continue
		}
		if _, ok = roles[anyUser]; ok {
			curRole = role
			break
		}
		if _, ok = roles[username]; ok {
			curRole = role
			break
		}
	}
	return curRole
}

func (t *TBotOpenAI) isBanned(username string) bool {
	_, ok := t.blacklist.Load(username)
	return username == "" || ok
}
