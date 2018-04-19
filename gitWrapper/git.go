package gitWrapper

import (
	"fmt"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/storer"
)

// Git struct wrapps Repository class from go-git to add a tag map used to perform queries when describing.
type Git struct {
	TagsMap map[plumbing.Hash]*plumbing.Reference
	*git.Repository
}

// PlainOpen opens a git repository from the given path. It detects if the
// repository is bare or a normal one. If the path doesn't contain a valid
// repository ErrRepositoryNotExists is returned
func PlainOpen(path string) (*Git, error) {
	r, err := git.PlainOpen(path)
	return &Git{
		make(map[plumbing.Hash]*plumbing.Reference),
		r,
	}, err
}

func (g *Git) getTagMap() error {
	tags, err := g.Tags()
	if err != nil {
		return err
	}

	err = tags.ForEach(func(t *plumbing.Reference) error {
		g.TagsMap[t.Hash()] = t
		return nil
	})

	return err
}

// Describe the reference as 'git describe --tags' will do
func (g *Git) Describe(reference *plumbing.Reference) (string, error) {

	// Fetch the reference log
	cIter, err := g.Log(&git.LogOptions{
		From:  reference.Hash(),
		Order: git.LogOrderCommitterTime,
	})

	// Build the tag map
	err = g.getTagMap()
	if err != nil {
		return "", err
	}

	// Search the tag
	var tag *plumbing.Reference
	var count int
	err = cIter.ForEach(func(c *object.Commit) error {
		if t, ok := g.TagsMap[c.Hash]; ok {
			tag = t
		}
		if tag != nil {
			return storer.ErrStop
		}
		count++
		return nil
	})
	if count == 0 {
		return fmt.Sprint(tag.Name().Short()), nil
	} else {
		return fmt.Sprintf("%v-%v-%v",
			tag.Name().Short(),
			count,
			tag.Hash().String()[0:8],
		), nil
	}
}
