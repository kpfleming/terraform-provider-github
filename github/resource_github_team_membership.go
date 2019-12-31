package github

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/google/go-github/v28/github"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceGithubTeamMembership() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubTeamMembershipCreateOrUpdate,
		Read:   resourceGithubTeamMembershipRead,
		Update: resourceGithubTeamMembershipCreateOrUpdate,
		Delete: resourceGithubTeamMembershipDelete,
		Importer: &schema.ResourceImporter{
			State: resourceGithubTeamMembershipImport,
		},

		Schema: map[string]*schema.Schema{
			"team_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateNumericIDFunc,
			},
			"user_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateNumericIDFunc,
			},
			"role": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "member",
				ValidateFunc: validateValueFunc([]string{"member", "maintainer"}),
			},
			"username": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceGithubTeamMembershipCreateOrUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Organization).client
	ctx := prepareResourceContext(d)

	teamIDString := d.Get("team_id").(string)
	userIDString := d.Get("user_id").(string)
	role := d.Get("role").(string)

	log.Printf("[DEBUG] Creating team membership: %s/%s (%s)", teamIDString, userIDString, role)

	teamID, _, username, err := getTeamAndUser(teamIDString, userIDString, meta.(*Organization))
	if err != nil {
		return err
	}

	_, _, err = client.Teams.AddTeamMembership(ctx,
		teamID,
		username,
		&github.TeamAddTeamMembershipOptions{
			Role: role,
		},
	)
	if err != nil {
		return err
	}

	d.SetId(buildTwoPartID(&teamIDString, &userIDString))

	return resourceGithubTeamMembershipRead(d, meta)
}

func resourceGithubTeamMembershipRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Organization).client
	ctx := prepareResourceContext(d)

	teamIDString, userIDString, err := parseTwoPartID(d.Id())
	if err != nil {
		return err
	}

	teamID, userID, username, err := getTeamAndUser(teamIDString, userIDString, meta.(*Organization))
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Reading team membership: %s/%s", teamIDString, userIDString)

	membership, resp, err := client.Teams.GetTeamMembership(ctx, teamID, username)

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[WARN] Removing team membership %s from state because it no longer exists in GitHub",
			d.Id())
		d.SetId("")
		return nil
	}

	if resp.StatusCode != http.StatusNotModified {
		if err != nil {
			return err
		}

		d.Set("etag", resp.Header.Get("ETag"))
		d.Set("team_id", teamID)
		d.Set("user_id", userID)
		d.Set("username", username)
		d.Set("role", membership.Role)
	}

	return nil
}

func resourceGithubTeamMembershipDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Organization).client
	ctx := prepareResourceContext(d)

	teamIDString, userIDString, err := parseTwoPartID(d.Id())
	if err != nil {
		return err
	}

	teamID, _, username, err := getTeamAndUser(teamIDString, userIDString, meta.(*Organization))
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Deleting team membership: %s/%s", teamIDString, userIDString)
	_, err = client.Teams.RemoveTeamMembership(ctx, teamID, username)

	return err
}

func resourceGithubTeamMembershipImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*Organization).client
	ctx := prepareResourceContext(d)

	if err := checkOrganization(meta); err != nil {
		return nil, err
	}

	teamString, userString, err := parseTwoPartID(d.Id())
	if err != nil {
		return nil, err
	}

	var teamID int64
	var userID int64

	log.Printf("[DEBUG] Reading team: %s", teamString)
	// Attempt to parse the string as a numeric ID
	teamID, err = strconv.ParseInt(teamString, 10, 64)
	if err != nil {
		// It wasn't a numeric ID, try to use it as a slug
		team, _, err := client.Teams.GetTeamBySlug(ctx, meta.(*Organization).name, teamString)
		if err != nil {
			return nil, err
		}
		teamID = *team.ID
	}

	log.Printf("[DEBUG] Reading user: %s", userString)
	// Attempt to parse the string as a numeric ID
	userID, err = strconv.ParseInt(userString, 10, 64)
	if err != nil {
		// It wasn't a numeric ID, try to use it as a username
		user, _, err := client.Users.Get(ctx, userString)
		if err != nil {
			return nil, err
		}
		userID = *user.ID
	}

	teamIDString := strconv.FormatInt(teamID, 10)
	userIDString := strconv.FormatInt(userID, 10)
	d.SetId(buildTwoPartID(&teamIDString, &userIDString))

	return []*schema.ResourceData{d}, nil
}

func getTeamAndUser(teamIDString string, userIDString string, org *Organization) (int64, int64, string, error) {
	teamID, err := strconv.ParseInt(teamIDString, 10, 64)
	if err != nil {
		return 0, 0, "", unconvertibleIdErr(teamIDString, err)
	}

	userID, err := strconv.ParseInt(userIDString, 10, 64)
	if err != nil {
		return 0, 0, "", unconvertibleIdErr(userIDString, err)
	}
	username, ok := org.UserMap.GetUsername(userID, org.client)
	if !ok {
		log.Printf("[DEBUG] Unable to obtain user %d from cache", userID)
		return 0, 0, "", fmt.Errorf("Unable to get GitHub user %d", userID)
	}

	return teamID, userID, username, nil
}
