package repo

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/SierraSoftworks/git-tool/internal/pkg/autocomplete"
	"github.com/SierraSoftworks/git-tool/internal/pkg/di"
	"github.com/SierraSoftworks/git-tool/internal/pkg/templates"
	"github.com/SierraSoftworks/git-tool/pkg/models"

	"github.com/sirupsen/logrus"

	"github.com/pkg/errors"
)

// A Mapper holds the information about a developer's source code folder which
// may contain multiple repositories.
type Mapper struct {
}

// GetBestRepo gets the repo which best matches a given name
func (d *Mapper) GetBestRepo(name string) (models.Repo, error) {
	if a := di.GetConfig().GetAlias(name); a != "" {
		name = a
	}

	r, err := d.GetRepo(name)
	if err != nil {
		return r, err
	}

	if r != nil {
		return r, nil
	}

	rs, err := d.GetRepos()
	if err != nil {
		return nil, err
	}

	matched := []models.Repo{}

	for _, rr := range rs {
		if autocomplete.Matches(templates.RepoQualifiedName(rr), name) {
			matched = append(matched, rr)
		}
	}

	if len(matched) == 1 {
		return matched[0], nil
	}

	return nil, nil
}

// GetRepos will fetch all of the repositories contained within a developer's dev
// directory which match the required naming scheme.
func (d *Mapper) GetRepos() ([]models.Repo, error) {
	logrus.WithField("path", di.GetConfig().DevelopmentDirectory()).Debug("Searching for repositories")

	files, err := ioutil.ReadDir(di.GetConfig().DevelopmentDirectory())
	if err != nil {
		return nil, errors.Wrapf(err, "repo: unable to list directory contents in dev directory '%s'", di.GetConfig().DevelopmentDirectory())
	}

	repos := []models.Repo{}

	for _, f := range files {
		if !f.IsDir() {
			continue
		}

		if f.Name() == "scratch" {
			continue
		}

		service := di.GetConfig().GetService(f.Name())
		if service == nil {
			logrus.WithField("service", f.Name()).Warn("Could not find a matching service entry in your configuration")
			continue
		}

		childRepos, err := d.GetReposForService(service)
		if err != nil {
			return nil, errors.Wrapf(err, "repo: unable to list directory contents in service directory '%s'", di.GetConfig().DevelopmentDirectory())
		}

		repos = append(repos, childRepos...)
	}

	return repos, nil
}

// GetScratchpads will fetch all of the known scratchpads which are stored locally.
func (d *Mapper) GetScratchpads() ([]models.Scratchpad, error) {
	logrus.Debug("Enumerating scratchpads")

	files, err := ioutil.ReadDir(di.GetConfig().ScratchDirectory())
	if err != nil {
		return nil, errors.Wrapf(err, "repo: unable to list directory contents in scratchpad direction '%s'", di.GetConfig().ScratchDirectory())
	}

	scratchpads := []models.Scratchpad{}

	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		scratchpad, err := d.GetScratchpad(file.Name())
		if err != nil {
			return nil, errors.Wrapf(err, "repo: failed to list directory contents in the scratchpad directory '%s'", di.GetConfig().ScratchDirectory())
		}
		scratchpads = append(scratchpads, scratchpad)
	}

	return scratchpads, nil
}

// GetScratchpad will fetch a scratchpad repo with the provided name
func (d *Mapper) GetScratchpad(name string) (models.Scratchpad, error) {
	return &scratchpad{
		fullName: name,
		path:     filepath.Join(di.GetConfig().ScratchDirectory(), name),
	}, nil
}

// GetReposForService will fetch all of the known repositories for a specific service.
func (d *Mapper) GetReposForService(service models.Service) ([]models.Repo, error) {
	logrus.WithField("service", service.Domain()).Debug("Enumerating repositories for service")

	path := filepath.Join(di.GetConfig().DevelopmentDirectory(), service.Domain())

	pattern := filepath.Join(path, service.DirectoryGlob())

	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, errors.Wrapf(err, "repo: unable to list directory contents in service directory '%s'", pattern)
	}

	repos := []models.Repo{}
	for _, f := range files {
		logrus.WithField("service", service.Domain()).WithField("path", f).Debug("Enumerated possible repository")
		r := &repo{
			service:  service,
			fullName: strings.Trim(strings.Replace(f[len(path):], string(filepath.Separator), "/", -1), "/"),
			path:     f,
		}

		if r.Exists() {
			repos = append(repos, r)
		} else {
			logrus.WithField("service", service.Domain()).WithField("path", f).Debug("Marked repository as invalid")
		}
	}

	return repos, nil
}

// GetRepo attempts to resolve the details of a repository given its name.
func (d *Mapper) GetRepo(name string) (models.Repo, error) {
	if name == "" {
		return d.GetCurrentDirectoryRepo()
	}

	dirParts := strings.Split(filepath.ToSlash(name), "/")
	if len(dirParts) < 2 {
		logrus.WithField("path", name).Debug("Not a fully qualified repository name")
		return nil, nil
	}

	serviceName := dirParts[0]
	service := di.GetConfig().GetService(serviceName)

	if service != nil {
		return d.GetRepoForService(service, filepath.Join(dirParts[1:]...))
	}

	r, err := d.GetFullyQualifiedRepo(name)
	if err != nil {
		return r, err
	}

	if r == nil {
		r, err = d.GetRepoForService(di.GetConfig().GetDefaultService(), name)
		if r != nil {
			if r.FullName() != filepath.ToSlash(name) {
				logrus.WithField("fullName", r.FullName()).WithField("name", name).Debug("Repo full name didn't match provided name")
				return nil, nil
			}

			return r, err
		}
	}

	logrus.WithField("path", name).Debug("Could not find a matching repository")
	return nil, nil
}

// GetRepoForService fetches the repo details for the named repository managed by the
// provided service.
func (d *Mapper) GetRepoForService(service models.Service, name string) (models.Repo, error) {
	dirParts := strings.Split(filepath.ToSlash(name), "/")

	fullNameLength := len(strings.Split(service.DirectoryGlob(), "/"))
	if len(dirParts) < fullNameLength {
		logrus.WithField("path", name).Debug("Not a fully named repository folder within the service's development directory")
		return nil, nil
	}

	r := NewRepo(service, strings.Join(dirParts[:fullNameLength], "/"))

	return r, nil
}

// GetFullyQualifiedRepo fetches the repo details for the fully qualified named
// repository which has been provided.
func (d *Mapper) GetFullyQualifiedRepo(name string) (models.Repo, error) {
	dirParts := strings.Split(filepath.ToSlash(name), "/")

	if len(dirParts) < 2 {
		// Not within a service's repository
		logrus.WithField("path", name).Debug("Not a repository folder within the development directory")
		return nil, nil
	}

	serviceName := dirParts[0]
	service := di.GetConfig().GetService(serviceName)
	if service == nil {
		logrus.WithField("path", name).Debug("No service found to handle repository type")
		return nil, nil
	}

	r, err := d.GetRepoForService(service, strings.Join(dirParts[1:], "/"))
	if err != nil {
		return r, err
	}

	return r, err
}

// GetCurrentDirectoryRepo gets the repo details for the repository open in your
// current directory.
func (d *Mapper) GetCurrentDirectoryRepo() (models.Repo, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, errors.Wrap(err, "repo: failed to get current directory")
	}

	devDirectory, err := filepath.EvalSymlinks(di.GetConfig().DevelopmentDirectory())
	if err != nil {
		devDirectory = di.GetConfig().DevelopmentDirectory()
	}

	if !d.inDevDirectory(devDirectory, dir) {
		logrus.WithField("path", dir).WithField("devdir", devDirectory).Debug("Not within the development directory")
		return nil, nil
	}

	localDir := strings.Trim(filepath.ToSlash(dir[len(devDirectory):]), "/")
	return d.GetFullyQualifiedRepo(localDir)
}

func (d *Mapper) inDevDirectory(devDirectory, path string) bool {
	// Quick check for guaranteed misses
	if !strings.HasPrefix(strings.ToLower(path), strings.ToLower(devDirectory)) {
		logrus.WithField("devdir", devDirectory).WithField("path", path).Debug("Dev directory match failed case-insensitive comparison")
		return false
	}

	devStat, err := os.Stat(devDirectory)
	if err != nil {
		if os.IsNotExist(err) {
			logrus.WithField("path", devDirectory).Error("Development directory does not exist")
			return false
		}

		logrus.WithError(err).Debug("Failed to stat development directory")
		return false
	}

	pathStat, err := os.Stat(path[:len(devDirectory)])
	if err != nil {
		logrus.WithField("path", path[:len(devDirectory)]).WithError(err).Debug("Failed to open dev directory in repo path")
		return false
	}

	logrus.WithField("devdir", devDirectory).WithField("path", path[:len(devDirectory)]).Debug("Comparing dev directory paths using os.SameFile")
	return os.SameFile(devStat, pathStat)
}
