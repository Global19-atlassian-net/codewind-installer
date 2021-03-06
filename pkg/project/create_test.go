/*******************************************************************************
 * Copyright (c) 2019 IBM Corporation and others.
 * All rights reserved. This program and the accompanying materials
 * are made available under the terms of the Eclipse Public License v2.0
 * which accompanies this distribution, and is available at
 * http://www.eclipse.org/legal/epl-v20.html
 *
 * Contributors:
 *     IBM Corporation - initial API and implementation
 *******************************************************************************/

package project

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/eclipse/codewind-installer/pkg/connections"
	"github.com/eclipse/codewind-installer/pkg/security"
	"github.com/eclipse/codewind-installer/pkg/test"
	"github.com/eclipse/codewind-installer/pkg/utils"
	"github.com/stretchr/testify/assert"
)

const testDir = "./testDir"

func TestDownloadTemplate(t *testing.T) {
	t.Run("success case: download insecure template", func(t *testing.T) {
		os.RemoveAll(testDir)
		defer os.RemoveAll(testDir)

		dest := filepath.Join(testDir, "insecureTemplateRepo")
		url := test.PublicGHRepoURL

		out, err := DownloadTemplate(dest, url, nil)

		assert.Equal(t, "success", out.Status)
		assert.Nil(t, err)
	})
	t.Run("success case: download GHE template using good username-password", func(t *testing.T) {
		if !test.UsingOwnGHECredentials {
			t.Skip("skipping this test because you haven't set GitHub credentials needed for this test")
		}

		os.RemoveAll(testDir)
		defer os.RemoveAll(testDir)

		dest := filepath.Join(testDir, "secureTemplateRepoGoodCredentials")
		url := test.GHERepoURL
		gitCredentials := &utils.GitCredentials{
			Username: test.GHEUsername,
			Password: test.GHEPassword,
		}

		out, err := DownloadTemplate(dest, url, gitCredentials)

		assert.NotNil(t, out)
		assert.Nil(t, err)
	})
	t.Run("success case: download GHE template using good personalAccessToken", func(t *testing.T) {
		if !test.UsingOwnGHECredentials {
			t.Skip("skipping this test because you haven't set GitHub credentials needed for this test")
		}

		os.RemoveAll(testDir)
		defer os.RemoveAll(testDir)

		dest := filepath.Join(testDir, "secureTemplateRepoGoodCredentials")
		url := test.GHERepoURL
		gitCredentials := &utils.GitCredentials{
			PersonalAccessToken: test.GHEPersonalAccessToken,
		}

		out, err := DownloadTemplate(dest, url, gitCredentials)

		assert.NotNil(t, out)
		assert.Nil(t, err)
	})
	t.Run("fail case: download GHE template using bad password", func(t *testing.T) {
		os.RemoveAll(testDir)
		defer os.RemoveAll(testDir)

		dest := filepath.Join(testDir, "secureTemplateRepoBadCredentials")
		url := test.GHERepoURL
		gitCredentials := &utils.GitCredentials{
			Username: test.GHEUsername,
			Password: "badpassword",
		}

		out, err := DownloadTemplate(dest, url, gitCredentials)

		assert.Nil(t, out)
		assert.Equal(t, errOpInvalidCredentials, err.Op)
		assert.Equal(t, http.StatusText(http.StatusUnauthorized), err.Desc)
	})
	t.Run("fail case: download GHE template using bad personalAccessToken)", func(t *testing.T) {
		os.RemoveAll(testDir)
		defer os.RemoveAll(testDir)

		dest := filepath.Join(testDir, "secureTemplateRepoBadCredentials")
		url := test.GHERepoURL
		gitCredentials := &utils.GitCredentials{
			Username: test.GHEUsername,
			Password: "badpersonalaccesstoken",
		}

		out, err := DownloadTemplate(dest, url, gitCredentials)

		assert.Nil(t, out)
		assert.Equal(t, errOpInvalidCredentials, err.Op)
		assert.Equal(t, http.StatusText(http.StatusUnauthorized), err.Desc)
	})
}

func TestDetermineProjectInfo(t *testing.T) {
	tests := map[string]struct {
		in            string
		wantLanguage  string
		wantBuildType string
		wantedErr     error
	}{
		"success case: liberty project": {
			in:            path.Join("../..", "resources", "test", "liberty-project"),
			wantLanguage:  "java",
			wantBuildType: "liberty",
		},
		"success case: spring project": {
			in:            path.Join("../..", "resources", "test", "spring-project"),
			wantLanguage:  "java",
			wantBuildType: "spring",
		},
		"success case: node.js project": {
			in:            path.Join("../..", "resources", "test", "node-project"),
			wantLanguage:  "javascript",
			wantBuildType: "nodejs",
		},
		"success case: swift project": {
			in:            path.Join("../..", "resources", "test", "swift-project"),
			wantLanguage:  "swift",
			wantBuildType: "swift",
		},
		"success case: python project": {
			in:            path.Join("../..", "resources", "test", "python-project"),
			wantLanguage:  "python",
			wantBuildType: "docker",
		},
		"success case: go project": {
			in:            path.Join("../..", "resources", "test", "go-project"),
			wantLanguage:  "go",
			wantBuildType: "docker",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			gotLanguage, gotBuildType := determineProjectInfo(test.in)

			assert.Equal(t, test.wantLanguage, gotLanguage)
			assert.Equal(t, test.wantBuildType, gotBuildType)
		})
	}
}

func TestWriteNewCwSettings(t *testing.T) {
	defaultInternalDebugPort := ""
	tests := map[string]struct {
		inProjectPath    string
		inBuildType      string
		wantCwSettings   CWSettings
		mockIgnoredPaths []string
	}{
		"success case: node project": {
			inProjectPath: "../../resources/test/node-project/.cw-settings",
			inBuildType:   "nodejs",
			wantCwSettings: CWSettings{
				ContextRoot:       "",
				InternalPort:      "",
				HealthCheck:       "",
				IsHTTPS:           false,
				InternalDebugPort: &defaultInternalDebugPort,
				StatusPingTimeout: "",
			},
			mockIgnoredPaths: []string{"*/node_modules*"},
		},
		"success case: liberty project": {
			inProjectPath: "../../resources/test/liberty-project/.cw-settings",
			inBuildType:   "liberty",
			wantCwSettings: CWSettings{
				ContextRoot:       "",
				InternalPort:      "",
				HealthCheck:       "",
				IsHTTPS:           false,
				InternalDebugPort: &defaultInternalDebugPort,
				MavenProfiles:     []string{""},
				MavenProperties:   []string{""},
				StatusPingTimeout: "",
			},
			mockIgnoredPaths: []string{"/libertyrepocache.zip"},
		},
		"success case: spring project": {
			inProjectPath: "../../resources/test/spring-project/.cw-settings",
			inBuildType:   "spring",
			wantCwSettings: CWSettings{
				ContextRoot:       "",
				InternalPort:      "",
				HealthCheck:       "",
				IsHTTPS:           false,
				InternalDebugPort: &defaultInternalDebugPort,
				MavenProfiles:     []string{""},
				MavenProperties:   []string{""},
				StatusPingTimeout: "",
			},
			mockIgnoredPaths: []string{"/localm2cache.zip"},
		},
		"success case: swift project": {
			inProjectPath: "../../resources/test/swift-project/.cw-settings",
			inBuildType:   "swift",
			wantCwSettings: CWSettings{
				ContextRoot:       "",
				InternalPort:      "",
				HealthCheck:       "",
				IsHTTPS:           false,
				StatusPingTimeout: "",
			},
			mockIgnoredPaths: []string{".swift-version"},
		},
		"success case: python project": {
			inProjectPath: "../../resources/test/python-project/.cw-settings",
			inBuildType:   "docker",
			wantCwSettings: CWSettings{
				ContextRoot:       "",
				InternalPort:      "",
				HealthCheck:       "",
				IsHTTPS:           false,
				StatusPingTimeout: "",
			},
			mockIgnoredPaths: []string{"*/.DS_Store"},
		},
		"success case: go project": {
			inProjectPath: "../../resources/test/go-project/.cw-settings",
			inBuildType:   "docker",
			wantCwSettings: CWSettings{
				ContextRoot:       "",
				InternalPort:      "",
				HealthCheck:       "",
				IsHTTPS:           false,
				StatusPingTimeout: "",
			},
			mockIgnoredPaths: []string{"*/.DS_Store"},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			jsonIgnoredPaths, _ := json.Marshal(test.mockIgnoredPaths)
			body := ioutil.NopCloser(bytes.NewReader([]byte(jsonIgnoredPaths)))
			mockClient := &security.ClientMockAuthenticate{StatusCode: http.StatusOK, Body: body}
			mockConnection := connections.Connection{ID: "local"}

			err := writeNewCwSettings(mockClient, &mockConnection, "dummyURL", test.inProjectPath, test.inBuildType)
			if err != nil {
				t.Errorf("writeNewCwSettings() returned error %s", err)
			}

			cwSettings := readCwSettings(test.inProjectPath)

			assert.Equal(t, test.wantCwSettings.ContextRoot, cwSettings.ContextRoot)
			assert.Equal(t, test.wantCwSettings.InternalPort, cwSettings.InternalPort)
			assert.Equal(t, test.wantCwSettings.HealthCheck, cwSettings.HealthCheck)
			assert.Equal(t, test.wantCwSettings.IsHTTPS, cwSettings.IsHTTPS)
			assert.Equal(t, test.wantCwSettings.StatusPingTimeout, cwSettings.StatusPingTimeout)
			assert.Equal(t, test.mockIgnoredPaths, cwSettings.IgnoredPaths)

			if test.wantCwSettings.InternalDebugPort != nil {
				assert.Equal(t, test.wantCwSettings.InternalDebugPort, cwSettings.InternalDebugPort)
			}
			if test.wantCwSettings.MavenProfiles != nil {
				assert.Equal(t, test.wantCwSettings.MavenProfiles, cwSettings.MavenProfiles)
			}
			if test.wantCwSettings.MavenProperties != nil {
				assert.Equal(t, test.wantCwSettings.MavenProperties, cwSettings.MavenProperties)
			}
			os.Remove(test.inProjectPath)
		})
	}
}

func TestProjectPathExists(t *testing.T) {
	errNoPath := errors.New(textNoProjectPath)
	errNoProject := errors.New(textProjectPathDoesNotExist)
	tests := map[string]struct {
		path      string
		wantError *ProjectError
	}{
		"success case - path exists": {
			path:      "../../resources/test/go-project",
			wantError: nil,
		},
		"error case - empty path": {
			path:      "",
			wantError: &ProjectError{errOpCreateProject, errNoPath, errNoPath.Error()},
		},
		"error case - no project at path": {
			path:      "not_a_path",
			wantError: &ProjectError{errOpCreateProject, errNoProject, errNoProject.Error()},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := checkProjectPathExists(test.path)
			assert.Equal(t, test.wantError, err)
		})
	}
}

func TestCheckProjectDirIsEmpty(t *testing.T) {
	errNoPath := errors.New(textNoProjectPath)
	errProjectNonEmpty := errors.New(textProjectPathNonEmpty)

	testFolder := "check_project_dir_empty_folder_delete_me"
	os.Mkdir(testFolder, 0777)
	tests := map[string]struct {
		path      string
		wantError *ProjectError
	}{
		"error case - empty path": {
			path:      "",
			wantError: &ProjectError{errOpCreateProject, errNoPath, errNoPath.Error()},
		},
		"error case - non empty path": {
			path:      "../../resources/test/go-project",
			wantError: &ProjectError{errOpCreateProject, errProjectNonEmpty, errProjectNonEmpty.Error()},
		},
		"success case - empty project at path": {
			path:      testFolder,
			wantError: nil,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := checkProjectDirIsEmpty(test.path)
			assert.Equal(t, test.wantError, err)
		})
	}
	os.RemoveAll(testFolder)
}

func TestRenameLegacySettings(t *testing.T) {
	testFolder := "rename_legacy_settings_delete_me"
	os.Mkdir(testFolder, 0777)
	legacySettingsPath := path.Join(testFolder, ".mc-settings")
	newSettingsPath := path.Join(testFolder, ".cw-settings")
	ioutil.WriteFile(legacySettingsPath, []byte{}, 0644)

	t.Run("error case - path to legacy settings does not exist", func(t *testing.T) {
		err := renameLegacySettings("/not_a_path/.mc-settings", "/not_a_path/.cw-settings")
		assert.Error(t, err)
	})

	t.Run("success case - renames legacy settings", func(t *testing.T) {
		err := renameLegacySettings(legacySettingsPath, newSettingsPath)
		assert.Nil(t, err)
		newSettingsFileExists := utils.PathExists(newSettingsPath)
		legacySettingsFileExists := utils.PathExists(legacySettingsPath)
		assert.True(t, newSettingsFileExists)
		assert.False(t, legacySettingsFileExists)
	})
	os.RemoveAll(testFolder)
}

func readCwSettings(filepath string) CWSettings {
	cwSettingsFile, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Println(err)
		return CWSettings{}
	}
	var cwSettings CWSettings
	json.Unmarshal(cwSettingsFile, &cwSettings)
	return cwSettings
}
