package internal

import (
	"testing"
	"github.com/javimosch/superlandings-go/internal/config"
	"github.com/javimosch/superlandings-go/internal/db"
	"github.com/javimosch/superlandings-go/internal/services"
)

func TestDomainResolution(t *testing.T) {
	cfg, _ := config.Load()
	db.InitDB(cfg.DBPath)
	dnsSvc := services.NewDNSService(cfg)
	domain, err := dnsSvc.GetDomainByDomain("vdb.dk2.intrane.fr")
	t.Logf("domain=%+v, err=%v\n", domain, err)
	
	siteSvc := services.NewSiteService(cfg)
	_, err = siteSvc.GetBySlug("")
	t.Logf("GetBySlug('') err=%v\n", err)
}
