package git

import (
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func OpenRepository(path string) (*git.Repository, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository at path '%s': %w", path, err)
	}

	return repo, nil
}

func GetHeadPatch(repo *git.Repository) (string, error) {
	headRef, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD ref: %w", err)
	}

	commit, err := repo.CommitObject(headRef.Hash())
	if err != nil {
		return "", fmt.Errorf("failed to get commit object: %w", err)
	}

	var parent *object.Commit
	if commit.NumParents() > 0 {
		parent, err = commit.Parent(0)
		if err != nil {
			return "", fmt.Errorf("failed to get parent commit: %w", err)
		}
	}

	patch, err := commit.Patch(parent)
	if err != nil {
		return "", fmt.Errorf("failed to generate patch: %w", err)
	}

	return patch.String(), nil
}

func GetCommitPatches(repo *git.Repository) (<-chan string, error) {
	cIter, err := repo.Log(&git.LogOptions{
		Order: git.LogOrderCommitterTime,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get commit log: %w", err)
	}

	patchChannel := make(chan string)

	go func() {
		defer close(patchChannel)

		cIter.ForEach(func(c *object.Commit) error {
			var parent *object.Commit
			if c.NumParents() > 0 {
				parent, _ = c.Parent(0)
			}

			patch, err := c.Patch(parent)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to get patch for commit %s: %v\n", c.Hash, err)
				return nil
			}

			patchChannel <- patch.String()
			return nil
		})
	}()
	return patchChannel, nil
}
