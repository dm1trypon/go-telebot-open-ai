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
	t.userRoles = map[string]map[string]struct{}{
		roleAdmin: admins,
		roleUser:  users,
	}
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
	t.permissions = map[string]map[string]struct{}{
		roleAdmin: adminCommands,
		roleUser:  userCommands,
	}
}

func (t *TBotOpenAI) checkPermissions(command, username string) bool {
	curRole := t.getRole(username)
	if curRole == "" {
		return false
	}
	if _, ok := t.permissions[curRole][anyUser]; ok {
		return true
	}
	if _, ok := t.permissions[curRole][command]; ok {
		return true
	}
	return false
}

func (t *TBotOpenAI) getRole(username string) string {
	var curRole string
	for _, role := range []string{roleAdmin, roleUser} {
		if _, ok := t.userRoles[role][anyUser]; ok {
			curRole = role
			break
		}
		if _, ok := t.userRoles[role][username]; ok {
			curRole = role
			break
		}
	}
	return curRole
}
