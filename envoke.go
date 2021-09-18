package main

import (
	. "github.com/watergist/file-engine/reader/structures"
	"log"
	"os"
	"path"
)

type nestedMapList map[string][]map[string]interface{}

func main() {
	Envoke()
}
func Envoke() {
	if len(os.Args) == 1 {
		log.Fatal("Where is the file to be processed?")
	}
	targetFile := os.Args[1]

	jsonLoaded := LoadJsonFile(targetFile+".json", &(nestedMapList{}))
	err := os.MkdirAll(targetFile, 0744)
	if err != nil {
		log.Fatal(err)
	}

	cache := map[string]interface{}{}
	arr := (*(jsonLoaded.(*nestedMapList)))["configs"]
	bootStrap := SelectKey(&arr, "@type", "type.googleapis.com/envoy.admin.v3.BootstrapConfigDump")
	if bootStrap != nil {
		WriteYaml(path.Join(targetFile, "bootstrap.yaml"), processBootstrap(bootStrap, &cache))
	}

	listeners := SelectKey(&arr, "@type", "type.googleapis.com/envoy.admin.v3.ListenersConfigDump")
	if listeners != nil {
		WriteYaml(path.Join(targetFile, "listeners.yaml"),
			map[string]interface{}{"resources": convertListeners(listeners)})
	}

	clusters := SelectKey(&arr, "@type", "type.googleapis.com/envoy.admin.v3.ClustersConfigDump")
	if clusters != nil {
		WriteYaml(path.Join(targetFile, "clusters.yaml"),
			map[string]interface{}{"resources": convertClusters(clusters)})
	}

	//secrets := SelectKey(&arr, "@type", "type.googleapis.com/envoy.admin.v3.SecretsConfigDump")
	//if secrets != nil {
	//	WriteYaml(path.Join(targetFile, "secrets.yaml"), (*secrets)["dynamic_secrets"])
	//}
	//
	//routes := SelectKey(&arr, "@type", "type.googleapis.com/envoy.admin.v3.RoutesConfigDump")
	//if routes != nil {
	//	WriteYaml(path.Join(targetFile, "routes.yaml"), (*routes)["dynamic_route_configs"])
	//}
	//
	//scopedRoutes := SelectKey(&arr, "@type", "type.googleapis.com/envoy.admin.v3.ScopedRoutesConfigDump")
	//if scopedRoutes != nil {
	//	WriteYaml(path.Join(targetFile, "scopedRoutes.yaml"), (*scopedRoutes)["dynamic_scoped_route_configs"])
	//}
	////save cache
	//WriteYaml(path.Join(targetFile, "cache.yaml"), cache)
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

func convertListeners(clusters *map[string]interface{}) (newClusters []interface{}) {
	if v := (*clusters)["dynamic_listeners"]; v != nil {
		for _, k := range v.([]interface{}) {
			if v := k.(map[string]interface{})["active_state"]; v != nil {
				if v := v.(map[string]interface{})["listener"]; v != nil {
					newClusters = append(newClusters, v)
				}
			}
		}
	}
	return newClusters
}

func processBootstrap(bootStrap *map[string]interface{}, cache *map[string]interface{}) (bootStrapNew *map[string]interface{}) {
	bootstrapConfig := (*bootStrap)["bootstrap"].(map[string]interface{})
	(*cache)["dynamic_resources"] =
		bootstrapConfig["dynamic_resources"]

	node := bootstrapConfig["node"].(map[string]interface{})
	delete(node, "hidden_envoy_deprecated_build_version")
	delete(node, "extensions")
	(*cache)["user_agent_build_version"] = node["user_agent_build_version"]
	delete(node, "user_agent_build_version")

	bootstrapConfig["dynamic_resources"] = map[string]map[string]string{
		"cds_config": {"path": "./clusters.yaml"},
		"lds_config": {"path": "./listeners.yaml"},
	}
	return &bootstrapConfig
}
