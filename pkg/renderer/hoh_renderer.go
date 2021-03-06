package renderer

import (
	"bufio"
	"bytes"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"strings"

	"github.com/openshift/library-go/pkg/assets"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	yamlserializer "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// HoHRenderer is an implementation of the Renderer interface for hub-of-hubs scenario
type HoHRenderer struct {
	manifestFS embed.FS
	decoder    runtime.Decoder
}

// NewHoHRenderer create a HoHRenderer with given filesystem
func NewHoHRenderer(manifestFS embed.FS) Renderer {
	return &HoHRenderer{
		manifestFS: manifestFS,
		decoder:    yamlserializer.NewDecodingSerializer(unstructured.UnstructuredJSONScheme),
	}
}

func (r *HoHRenderer) Render(component string, getConfigValuesFunc GetConfigValuesFunc) ([]runtime.Object, error) {
	return r.RenderWithFilter(component, "", getConfigValuesFunc)
}

func (r *HoHRenderer) RenderWithFilter(component, filter string, getConfigValuesFunc GetConfigValuesFunc) ([]runtime.Object, error) {
	var objects []runtime.Object

	configValues, err := getConfigValuesFunc(component)
	if err != nil {
		return objects, err
	}

	templateFiles, err := getTemplateFiles(r.manifestFS, component, filter)
	if err != nil {
		return objects, err
	}
	if len(templateFiles) == 0 {
		return objects, fmt.Errorf("no template files found")
	}

	for _, template := range templateFiles {
		templateContent, err := r.manifestFS.ReadFile(template)
		if err != nil {
			return objects, err
		}

		if len(templateContent) == 0 {
			continue
		}

		raw := assets.MustCreateAssetFromTemplate(template, templateContent, configValues).Data
		yamlReader := yaml.NewYAMLReader(bufio.NewReader(bytes.NewReader(raw)))
		for {
			b, err := yamlReader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				return objects, err
			}
			if len(b) != 0 {
				object, _, err := r.decoder.Decode(b, nil, nil)
				if err != nil {
					if runtime.IsMissingKind(err) {
						continue
					}
					return objects, err
				}
				objects = append(objects, object)
			}
		}
	}

	return objects, nil
}

func (r *HoHRenderer) RenderForCluster(cluster, component string, getClusterConfigValuesFunc GetClusterConfigValuesFunc) ([]runtime.Object, error) {
	return r.RenderForClusterWithFilter(cluster, component, "", getClusterConfigValuesFunc)
}

func (r *HoHRenderer) RenderForClusterWithFilter(cluster, component, filter string, getClusterConfigValuesFunc GetClusterConfigValuesFunc) ([]runtime.Object, error) {
	var objects []runtime.Object

	configValues, err := getClusterConfigValuesFunc(cluster, component)
	if err != nil {
		return objects, err
	}

	templateFiles, err := getTemplateFiles(r.manifestFS, component, filter)
	if err != nil {
		return objects, err
	}
	if len(templateFiles) == 0 {
		return objects, fmt.Errorf("no template files found")
	}

	for _, template := range templateFiles {
		templateContent, err := r.manifestFS.ReadFile(template)
		if err != nil {
			return objects, err
		}

		if len(templateContent) == 0 {
			continue
		}

		raw := assets.MustCreateAssetFromTemplate(template, templateContent, configValues).Data
		object, _, err := r.decoder.Decode(raw, nil, nil)
		if err != nil {
			if runtime.IsMissingKind(err) {
				continue
			}
			return objects, err
		}
		objects = append(objects, object)
	}

	return objects, nil
}

func getTemplateFiles(manifestFS embed.FS, dir, filter string) ([]string, error) {
	files, err := getFiles(manifestFS)
	if err != nil {
		return nil, err
	}
	if dir == "." || len(dir) == 0 {
		return files, nil
	}

	var templateFiles []string
	for _, file := range files {
		if strings.HasPrefix(file, dir) && strings.Contains(file, filter) {
			templateFiles = append(templateFiles, file)
		}
	}

	return templateFiles, nil
}

func getFiles(manifestFS embed.FS) ([]string, error) {
	var files []string
	err := fs.WalkDir(manifestFS, ".", func(file string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		files = append(files, file)
		return nil
	})
	return files, err
}
