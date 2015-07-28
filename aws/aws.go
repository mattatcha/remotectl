package aws

import (
	"errors"
	"log"

	env "github.com/MattAitchison/envconfig"
	"github.com/MattAitchison/remotectl/providers"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

var region = env.String("aws_region", "", "aws ec2 region")

func init() {
	providers.Providers.Register(new(AWSProvider), "aws")
}

// AWSProvider is a provider for digitalocean
type AWSProvider struct {
	// svc *aws.EC2
	svc *ec2.EC2
}

// Setup will get the DO key and login.
func (p *AWSProvider) Setup() error {
	p.svc = ec2.New(&aws.Config{Region: region})
	// credentials.Get()
	//
	// doToken = strings.TrimSpace(doToken)
	// if len(doToken) == 0 {
	// 	log.Fatal("access key required")
	// }
	//
	// t := &oauth.Transport{
	// 	Token: &oauth.Token{AccessToken: doToken},
	// }
	return errors.New("aws provider setup error")

}

// Query will retrieve all digitalocean droplets
// Rename to Query with namespace and query string as args.
func (p *AWSProvider) Query(namespace, query string) ([]providers.Host, error) {
	// Get list of instances
	resp, err := p.svc.DescribeInstances(nil)
	if err != nil {
		panic(err)
	}

	log.Println(resp)

	hosts := []providers.Host{}

	// for _, drop := range drops {
	// 	if len(namespace) != 0 && !strings.HasPrefix(drop.Name, namespace) {
	// 		continue
	// 	}
	//
	// 	host := providers.Host{
	// 		Name: drop.Name,
	// 		Addr: getPublicIP(drop),
	// 	}
	//
	// 	hosts = append(hosts, host)
	// }

	return hosts, nil
}
