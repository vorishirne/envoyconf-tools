package main

import (
	. "github.com/watergist/file-engine/reader/structures"
	"log"
	"os"
	"path"
	"strings"
)

type nestedMapList map[string][]map[string]interface{}

func main() {
	if len(os.Args) == 1 {
		log.Fatal("Where is the file to be processed?")
	}
	for i := 1; i < len(os.Args); i++ {
		Envoke(os.Args[i])
	}
}

func Envoke(targetFile string) {
	targetFile = strings.TrimSpace(targetFile)
	if !strings.HasSuffix(targetFile, ".json") {
		log.Fatalf("Why don't you give a json file instead of %v", targetFile)
	}
	jsonLoaded := LoadJsonFile(targetFile, &(nestedMapList{}))
	targetDir := strings.TrimSuffix(targetFile, ".json")

	err := os.MkdirAll(targetDir, 0744)
	if err != nil {
		log.Fatal(err)
	}

	// this is the original structure of config_dump
	arr := (*(jsonLoaded.(*nestedMapList)))["configs"]

	bootStrap := SelectKey(&arr, "@type", "type.googleapis.com/envoy.admin.v3.BootstrapConfigDump")
	if bootStrap != nil {
		WriteYaml(path.Join(targetDir, "bootstrap.yaml"), processBootstrap(bootStrap))
	}

	listeners := SelectKey(&arr, "@type", "type.googleapis.com/envoy.admin.v3.ListenersConfigDump")
	if listeners != nil {
		WriteYaml(path.Join(targetDir, "listeners_active.yaml"),
			map[string]interface{}{"resources": convertListeners(listeners)})
	}
	if listeners != nil {
		WriteYaml(path.Join(targetDir, "listeners_errored.yaml"),
			map[string]interface{}{"resources": convertErroredListeners(listeners)})
	}

	clusters := SelectKey(&arr, "@type", "type.googleapis.com/envoy.admin.v3.ClustersConfigDump")
	if clusters != nil {
		WriteYaml(path.Join(targetDir, "clusters.yaml"),
			map[string]interface{}{"resources": convertClusters(clusters)})
	}

	secrets := SelectKey(&arr, "@type", "type.googleapis.com/envoy.admin.v3.SecretsConfigDump")
	if secrets != nil {
		WriteYaml(path.Join(targetDir, "secrets.yaml"), map[string]interface{}{"resources": convertSecrets(secrets)})
	}

	routes := SelectKey(&arr, "@type", "type.googleapis.com/envoy.admin.v3.RoutesConfigDump")
	if routes != nil {
		WriteYaml(path.Join(targetDir, "routes.yaml"), map[string]interface{}{"resources": convertRoutes(routes)})
	}

	scopedRoutes := SelectKey(&arr, "@type", "type.googleapis.com/envoy.admin.v3.ScopedRoutesConfigDump")
	if scopedRoutes != nil {
		WriteYaml(path.Join(targetDir, "scopedRoutes.yaml"), map[string]interface{}{"resources": convertScopedRoutes(scopedRoutes)})
	}
	//save cache
	WriteYaml(path.Join(targetDir, "zConfig.yaml"), jsonLoaded)
}

func convertClusters(clusters *map[string]interface{}) (newClusters []interface{}) {
	if v := (*clusters)["dynamic_active_clusters"]; v != nil {
		for _, k := range v.([]interface{}) {
			if v := k.(map[string]interface{})["cluster"]; v != nil {
				newClusters = append(newClusters, v)
			}
		}
	}
	return newClusters
}

func convertListeners(listeners *map[string]interface{}) (newListeners []interface{}) {
	if v := (*listeners)["dynamic_listeners"]; v != nil {
		for _, k := range v.([]interface{}) {
			if v := k.(map[string]interface{})["active_state"]; v != nil {
				if v := v.(map[string]interface{})["listener"]; v != nil {
					newListeners = append(newListeners, v)
				}
			}
		}
	}
	return newListeners
}

func convertErroredListeners(listeners *map[string]interface{}) (newListeners []interface{}) {
	if v := (*listeners)["dynamic_listeners"]; v != nil {
		for _, k := range v.([]interface{}) {
			if v := k.(map[string]interface{})["error_state"]; v != nil {
				newListeners = append(newListeners, k)
			}
		}
	}
	return newListeners
}

func convertRoutes(routes *map[string]interface{}) (newRoutes []interface{}) {
	if v := (*routes)["dynamic_route_configs"]; v != nil {
		for _, k := range v.([]interface{}) {
			if v := k.(map[string]interface{})["route_config"]; v != nil {
				newRoutes = append(newRoutes, v)
			}
		}
	}
	return newRoutes
}

// this was never tested
func convertScopedRoutes(staticRoutes *map[string]interface{}) (newStaticRoutes []interface{}) {
	if v := (*staticRoutes)["dynamic_scoped_route_configs"]; v != nil {
		for _, k := range v.([]interface{}) {
			if v := k.(map[string]interface{})["static_route_config"]; v != nil {
				newStaticRoutes = append(newStaticRoutes, v)
			}
		}
	}
	return newStaticRoutes
}

func convertSecrets(secrets *map[string]interface{}) (newSecrets []interface{}) {
	if v := (*secrets)["dynamic_active_secrets"]; v != nil {
		for _, k := range v.([]interface{}) {
			if v := k.(map[string]interface{})["secret"]; v != nil {
				newSecrets = append(newSecrets, v)
			}
		}
	}
	return newSecrets
}

func processBootstrap(bootStrap *map[string]interface{}) (bootStrapNew *map[string]interface{}) {
	bootstrapConfig := (*bootStrap)["bootstrap"].(map[string]interface{})

	bootstrapConfig["dynamic_resources"] = map[string]map[string]string{
		"cds_config": {"path": "./clusters.yaml"},
		"lds_config": {"path": "./listeners_active.yaml"},
	}

	node := bootstrapConfig["node"].(map[string]interface{})
	delete(node, "hidden_envoy_deprecated_build_version")
	delete(node, "extensions")
	delete(node, "user_agent_build_version")

	return &bootstrapConfig
}

// prepare bootstrapConfig var which will have the entire bootstrap yaml
// first we get config_dump's bootstrap part in bootstrapConfig
// then update the dynamic_resources, and saving original one in cache
// then copy the node's hashmap (only it's reference) to a separate variable
// and use that to remove unwanted fields easily, instead of using original ds
