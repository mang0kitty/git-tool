package tasks

import (
	"github.com/SierraSoftworks/git-tool/pkg/models"
	"github.com/pkg/errors"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

func NewBranch(name, fromRef string) Task {
	return &gitNewBranch{
		BranchName: name,
		FromRef:    fromRef,
	}
}

type gitNewBranch struct {
	BranchName string
	FromRef    string
}

func (t *gitNewBranch) ApplyRepo(r models.Repo) error {
	gr, err := git.PlainOpen(r.Path())
	if err != nil {
		return errors.Wrap(err, "repo: unable to open git repository")
	}

	_, err = gr.Branch(t.BranchName)
	switch err {
	case git.ErrBranchNotFound:
		fromRef, err := gr.ResolveRevision(plumbing.Revision(t.FromRef))
		if err != nil {
			return errors.Wrapf(err, "repo: unable to resolve reference '%s'", t.FromRef)
		}

		ref := plumbing.NewHashReference(plumbing.NewBranchReferenceName(t.BranchName), *fromRef)
		return errors.Wrap(gr.Storer.SetReference(ref), "repo: failed to create new branch")
	case nil:
		return nil
	default:
		return errors.Wrap(err, "repo: unable to determine whether branch is already created")
	}

	if err != nil && err != git.ErrBranchNotFound {
		return errors.Wrap(err, "repo: unable to determine whether branch is already created")
	}

	if err != nil && err == git.ErrBranchNotFound {

	}

	return nil
}

func (t *gitNewBranch) ApplyScratchpad(s models.Scratchpad) error {
	return nil
}
