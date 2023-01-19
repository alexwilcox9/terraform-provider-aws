package opsworks

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceUserProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceUserProfileCreate,
		Read:   resourceUserProfileRead,
		Update: resourceUserProfileUpdate,
		Delete: resourceUserProfileDelete,

		Schema: map[string]*schema.Schema{
			"allow_self_management": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"ssh_public_key": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ssh_username": {
				Type:     schema.TypeString,
				Required: true,
			},
			"user_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceUserProfileCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).OpsWorksConn()

	iamUserARN := d.Get("user_arn").(string)
	input := &opsworks.CreateUserProfileInput{
		AllowSelfManagement: aws.Bool(d.Get("allow_self_management").(bool)),
		IamUserArn:          aws.String(iamUserARN),
		SshUsername:         aws.String(d.Get("ssh_username").(string)),
	}

	if v, ok := d.GetOk("ssh_public_key"); ok {
		input.SshPublicKey = aws.String(v.(string))
	}

	_, err := conn.CreateUserProfile(input)

	if err != nil {
		return fmt.Errorf("creating OpsWorks User Profile (%s): %w", iamUserARN, err)
	}

	d.SetId(iamUserARN)

	return resourceUserProfileUpdate(d, meta)
}

func resourceUserProfileRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).OpsWorksConn()

	profile, err := FindUserProfileByARN(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] OpsWorks User Profile %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading OpsWorks User Profile (%s): %w", d.Id(), err)
	}

	d.Set("allow_self_management", profile.AllowSelfManagement)
	d.Set("ssh_public_key", profile.SshPublicKey)
	d.Set("ssh_username", profile.SshUsername)
	d.Set("user_arn", profile.IamUserArn)

	return nil
}

func resourceUserProfileUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).OpsWorksConn()

	input := &opsworks.UpdateUserProfileInput{
		AllowSelfManagement: aws.Bool(d.Get("allow_self_management").(bool)),
		IamUserArn:          aws.String(d.Get("user_arn").(string)),
		SshPublicKey:        aws.String(d.Get("ssh_public_key").(string)),
		SshUsername:         aws.String(d.Get("ssh_username").(string)),
	}

	_, err := conn.UpdateUserProfile(input)

	if err != nil {
		return fmt.Errorf("updating OpsWorks User Profile (%s): %w", d.Id(), err)
	}

	return resourceUserProfileRead(d, meta)
}

func resourceUserProfileDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).OpsWorksConn()

	log.Printf("[DEBUG] Deleting OpsWorks User Profile: %s", d.Id())
	_, err := conn.DeleteUserProfile(&opsworks.DeleteUserProfileInput{
		IamUserArn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, opsworks.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting OpsWorks User Profile (%s): %w", d.Id(), err)
	}

	return nil
}

func FindUserProfileByARN(conn *opsworks.OpsWorks, arn string) (*opsworks.UserProfile, error) {
	input := &opsworks.DescribeUserProfilesInput{
		IamUserArns: aws.StringSlice([]string{arn}),
	}

	output, err := conn.DescribeUserProfiles(input)

	if tfawserr.ErrCodeEquals(err, opsworks.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.UserProfiles) == 0 || output.UserProfiles[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.UserProfiles); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.UserProfiles[0], nil
}
