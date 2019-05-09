package flows

import "github.com/privatix/dapp-installer/flow"

// Methods.
const (
	MethodInstall   = "install"
	MethodUpdate    = "update"
	MethodRemove    = "remove"
	MethodStart     = "start"
	MethodStop      = "stop"
	MethodRunV2Ray  = "run-v2ray"
	MethodRunPlugin = "run-plugin"
)

// Help messages.
const (
	RootHelp = `
Usage:
  installer [command] [flags]
Available Commands:
  install     Install product package
  remove      Remove product package
  run         Run service
  start	      Start service
  stop	      Stop service
  run-v2ray   Runs v2ray
  run-plugin  Runs dapp proxy plugin
Flags:
  --help      Display help information
Use "installer [command] --help" for more information about a command.
`
	commonHelpFormat = `
    Usage:
        installer %s [flags]
    Flags:
      --help    Display help information
      --role    Product role, either 'client' or 'agent'
      --proddir Product install directory
`
	helpRemove = `
Usage:
	installer remove [flags]
Flags:
  --help    Display help information
  --proddir Product install directory
`
	helpUpdate = `
	Usage:
		installer update [flags]
	Flags:
		--help    Display help information
		--proddir Product install directory
`
)

// Install is a proxy plug-in service installation flow.
func Install() flow.Flow {
	return flow.Flow{
		Name: "Proxy installation",
		Steps: []flow.Step{
			newStep("read & proccess flags for installation", parseInstallFlags, nil),
			newStep("validate install environment", validateInstallEnvironment, nil),
			newStep("prepare plugin configs", preparePluginConfigs, nil),
			newStep("create v2ray daemons", createV2RayDaemon, removeV2RayDaemon),
			newStep("create plugin daemons", createPluginDaemon, removePluginDaemon),
			newStep("start v2ray daemons", startV2rayDaemon, stopV2rayDaemon),
			newStep("start plugin daemons", startPluginDaemon, stopPluginDaemon),
			newStep("save installation details", saveInstallationDetails, nil),
		},
	}
}

// Remove is a proxy plug-in service remove flow.
func Remove() flow.Flow {
	return flow.Flow{
		Name: "Proxy remove",
		Steps: []flow.Step{
			newStep("read & proccess flags for remove", parseRemoveFlags, nil),
			newStep("read installation details", readInstallationDetails, nil),
			newStep("stop plugin daemon", stopPluginDaemon, startPluginDaemon),
			newStep("stop v2ray daemon", stopV2rayDaemon, startV2rayDaemon),
			newStep("remove v2ray daemon", removeV2RayDaemon, createV2RayDaemon),
			newStep("remove plugin daemon", removePluginDaemon, createPluginDaemon),
		},
	}
}

// Update is a proxy plug-in service update flow.
func Update() flow.Flow {
	return flow.Flow{
		Name: "Proxy update",
		Steps: []flow.Step{
			newStep("read & proccess flags for update", parseUpdateFlags, nil),
			newStep("read installation details", readInstallationDetails, nil),
			newStep("locate product temp dir", locateProductTempDir, nil),
			newStep("prepare plugin configs in product temp dir", prepareUpdateFromPluginConfigs, nil),
			newStep("copy /data files", copyDataDirFiles, nil),
			newStep("merge configs", copyAndMergeConfigs, nil),
		},
	}
}

// Start is a proxy plug-in service start flow.
func Start() flow.Flow {
	return flow.Flow{
		Name: "start proxy",
		Steps: []flow.Step{
			newStep("start v2ray daemon", startV2rayDaemon, stopV2rayDaemon),
			newStep("start plugin daemons", startPluginDaemon, stopPluginDaemon),
		},
	}
}

// Stop is a proxy plug-in service stop flow.
func Stop() flow.Flow {
	return flow.Flow{
		Name: "",
		Steps: []flow.Step{
			newStep("stop v2ray daemon", stopV2rayDaemon, startV2rayDaemon),
			newStep("stop plugin daemons", stopPluginDaemon, startPluginDaemon),
		},
	}
}

// RunV2Ray starts v2ray.
func RunV2Ray() flow.Flow {
	return flow.Flow{
		Name: "Run v2ray",
		Steps: []flow.Step{
			newStep("read & proccess flags for v2ray run", parseV2RayRunFlags, nil),
			newStep("run v2ray", runV2Ray, nil),
		},
	}
}

// RunPlugin starts plugin.
func RunPlugin() flow.Flow {
	return flow.Flow{
		Name: "Run proxy plug-in",
		Steps: []flow.Step{
			newStep("read & proccess flags for plugin run", parsePluginRunFlags, nil),
			newStep("run plugin", runPlugin, nil),
		},
	}
}
