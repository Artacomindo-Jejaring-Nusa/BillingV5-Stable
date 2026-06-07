package domain

import "context"

// MikrotikRepository defines database operations for MikrotikServer
type MikrotikRepository interface {
	GetAll(ctx context.Context) ([]MikrotikServer, error)
	GetByID(ctx context.Context, id uint64) (*MikrotikServer, error)
	Create(ctx context.Context, server *MikrotikServer) error
	Update(ctx context.Context, server *MikrotikServer) error
	Delete(ctx context.Context, id uint64) error
	GetByName(ctx context.Context, name string) (*MikrotikServer, error)
	GetByNames(ctx context.Context, names []string) ([]MikrotikServer, error)
}

// MikrotikUsecase defines business logic for MikrotikServer
type MikrotikUsecase interface {
	FetchAll(ctx context.Context) ([]MikrotikServer, error)
	GetByID(ctx context.Context, id uint64) (*MikrotikServer, error)
	Store(ctx context.Context, server *MikrotikServer) error
	Update(ctx context.Context, id uint64, server *MikrotikServer) error
	Delete(ctx context.Context, id uint64) error
	TestConnection(ctx context.Context, id uint64) (map[string]interface{}, *MikrotikServer, error)
}

// DataTeknisRepository defines database operations for DataTeknis
type DataTeknisRepository interface {
	GetAll(ctx context.Context, skip, limit int, search string, olt string, profile string, vlan string, onuPowerMin, onuPowerMax *int) ([]DataTeknis, int64, error)
	GetByID(ctx context.Context, id uint64) (*DataTeknis, error)
	GetByPelangganID(ctx context.Context, pelangganID uint64) (*DataTeknis, error)
	Create(ctx context.Context, data *DataTeknis) error
	Update(ctx context.Context, data *DataTeknis) error
	Delete(ctx context.Context, id uint64) error
	GetAvailableOLT(ctx context.Context) ([]string, error)
	GetAvailableProfiles(ctx context.Context) ([]string, error)
	GetAvailableVlans(ctx context.Context) ([]string, error)
	GetOnuPowerRanges(ctx context.Context) (*int, *int, error)
	CheckIPAddress(ctx context.Context, ip string, excludeID *uint64) (bool, error)
	GetOdpByCode(ctx context.Context, code string) (*ODP, error)
	GetOdpByCodes(ctx context.Context, codes []string) ([]ODP, error)
}

// DataTeknisUsecase defines business logic for DataTeknis
type DataTeknisUsecase interface {
	FetchAll(ctx context.Context, skip, limit int, search string, olt string, profile string, vlan string, onuPowerMin, onuPowerMax *int) ([]DataTeknis, int64, error)
	GetByID(ctx context.Context, id uint64) (*DataTeknis, error)
	GetByPelangganID(ctx context.Context, pelangganID uint64) (*DataTeknis, error)
	Store(ctx context.Context, data *DataTeknis) error
	Update(ctx context.Context, id uint64, data *DataTeknis) error
	Delete(ctx context.Context, id uint64) error
	GetAvailableOLT(ctx context.Context) ([]string, error)
	GetAvailableProfiles(ctx context.Context) ([]string, error)
	GetAvailableVlans(ctx context.Context) ([]string, error)
	GetOnuPowerRanges(ctx context.Context) (map[string]int, error)
	CheckIPAddress(ctx context.Context, ip string, excludeID *uint64) (bool, error)
	GetAvailableProfilesForPackage(ctx context.Context, packageID uint64, pelangganID uint64, mikrotikServerID *uint64) ([]map[string]interface{}, error)
	GetLastUsedIP(ctx context.Context, mikrotikServerID uint64) (map[string]interface{}, error)
	ImportFromCSV(ctx context.Context, csvContent string) (int, error)
	Export(ctx context.Context, format string) ([]byte, string, error)
}

// OLTRepository defines database operations for OLT
type OLTRepository interface {
	GetAll(ctx context.Context) ([]OLT, error)
	GetByID(ctx context.Context, id uint64) (*OLT, error)
	Create(ctx context.Context, olt *OLT) error
	Update(ctx context.Context, olt *OLT) error
	Delete(ctx context.Context, id uint64) error
}

// OLTUsecase defines business logic for OLT
type OLTUsecase interface {
	FetchAll(ctx context.Context) ([]OLT, error)
	GetByID(ctx context.Context, id uint64) (*OLT, error)
	Store(ctx context.Context, olt *OLT) error
	Update(ctx context.Context, id uint64, olt *OLT) error
	Delete(ctx context.Context, id uint64) error
	TestConnection(ctx context.Context, id uint64) (string, error)
}

// ODPRepository defines database operations for ODP
type ODPRepository interface {
	GetAll(ctx context.Context) ([]ODP, error)
	GetByID(ctx context.Context, id uint64) (*ODP, error)
	Create(ctx context.Context, odp *ODP) error
	Update(ctx context.Context, odp *ODP) error
	Delete(ctx context.Context, id uint64) error
}

// ODPUsecase defines business logic for ODP
type ODPUsecase interface {
	FetchAll(ctx context.Context) ([]ODP, error)
	GetByID(ctx context.Context, id uint64) (*ODP, error)
	Store(ctx context.Context, odp *ODP) error
	Update(ctx context.Context, id uint64, odp *ODP) error
	Delete(ctx context.Context, id uint64) error
}


