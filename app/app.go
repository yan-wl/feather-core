package app

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spf13/cast"

	tmdb "github.com/cometbft/cometbft-db"
	abcitypes "github.com/cometbft/cometbft/abci/types"
	tmjson "github.com/cometbft/cometbft/libs/json"
	"github.com/cometbft/cometbft/libs/log"
	tmos "github.com/cometbft/cometbft/libs/os"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsim "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/cosmos/cosmos-sdk/x/consensus"

	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"

	"github.com/cosmos/cosmos-sdk/x/bank"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/cosmos/cosmos-sdk/x/capability"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"

	consensuskeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"

	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"

	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	"github.com/cosmos/cosmos-sdk/x/evidence"
	evidencekeeper "github.com/cosmos/cosmos-sdk/x/evidence/keeper"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"

	"github.com/cosmos/cosmos-sdk/x/feegrant"
	feegrantkeeper "github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	feegrantmodule "github.com/cosmos/cosmos-sdk/x/feegrant/module"

	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	"github.com/cosmos/cosmos-sdk/x/gov"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/cosmos/cosmos-sdk/x/group"
	groupkeeper "github.com/cosmos/cosmos-sdk/x/group/keeper"
	groupmodule "github.com/cosmos/cosmos-sdk/x/group/module"

	"github.com/cosmos/cosmos-sdk/x/mint"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	"github.com/cosmos/cosmos-sdk/x/nft"
	nftkeeper "github.com/cosmos/cosmos-sdk/x/nft/keeper"
	nftmodule "github.com/cosmos/cosmos-sdk/x/nft/module"

	"github.com/cosmos/cosmos-sdk/x/params"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	paramsproposaltypes "github.com/cosmos/cosmos-sdk/x/params/types/proposal"

	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"

	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradeclient "github.com/cosmos/cosmos-sdk/x/upgrade/client"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	ica "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/keeper"
	icacontrollertypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/types"
	icahostkeeper "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/keeper"
	icahosttypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	ibctransfer "github.com/cosmos/ibc-go/v7/modules/apps/transfer"
	ibctransferkeeper "github.com/cosmos/ibc-go/v7/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v7/modules/core"
	ibcclient "github.com/cosmos/ibc-go/v7/modules/core/02-client"
	ibcclientclient "github.com/cosmos/ibc-go/v7/modules/core/02-client/client"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	alliancebank "github.com/terra-money/alliance/custom/bank"
	alliancebankkeeper "github.com/terra-money/alliance/custom/bank/keeper"
	alliance "github.com/terra-money/alliance/x/alliance"
	alliancekeeper "github.com/terra-money/alliance/x/alliance/keeper"
	alliancetypes "github.com/terra-money/alliance/x/alliance/types"

	"github.com/terra-money/feather-core/app/openapiconsole"
	appparams "github.com/terra-money/feather-core/app/params"
	"github.com/terra-money/feather-core/docs"
)

const (
	AccountAddressPrefix = "feath"
	Name                 = "feather-core"
)

// TODO: What is this?
func getGovProposalHandlers() []govclient.ProposalHandler {
	var govProposalHandlers []govclient.ProposalHandler
	// this line is used by starport scaffolding # stargate/app/govProposalHandlers

	govProposalHandlers = append(govProposalHandlers,
		paramsclient.ProposalHandler,
		upgradeclient.LegacyProposalHandler,
		upgradeclient.LegacyCancelProposalHandler,
		ibcclientclient.UpdateClientProposalHandler,
		ibcclientclient.UpgradeProposalHandler,
		// allianceclient.CreateAllianceProposalHandler,
		// allianceclient.UpdateAllianceProposalHandler,
		// allianceclient.DeleteAllianceProposalHandler,
		// this line is used by starport scaffolding # stargate/app/govProposalHandler
	)

	return govProposalHandlers
}

var (
	// DefaultNodeHome default home directories for the application daemon
	DefaultNodeHome string
)

var (
	_ servertypes.Application = (*App)(nil)
	_ runtime.AppI            = (*App)(nil)
)

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	DefaultNodeHome = filepath.Join(userHomeDir, "."+Name)
}

// App extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type App struct {
	*baseapp.BaseApp

	cdc               *codec.LegacyAmino
	appCodec          codec.Codec
	interfaceRegistry types.InterfaceRegistry

	invCheckPeriod uint

	// keys to access the substores
	keys    map[string]*storetypes.KVStoreKey // TODO: Is this state even needed?
	tkeys   map[string]*storetypes.TransientStoreKey
	memKeys map[string]*storetypes.MemoryStoreKey

	// keepers
	AuthKeeper            authkeeper.AccountKeeper // TODO: Do we even need to store this state?
	AuthzKeeper           authzkeeper.Keeper
	BankKeeper            alliancebankkeeper.Keeper
	CapabilityKeeper      *capabilitykeeper.Keeper
	StakingKeeper         *stakingkeeper.Keeper
	SlashingKeeper        slashingkeeper.Keeper
	MintKeeper            mintkeeper.Keeper
	DistrKeeper           distrkeeper.Keeper
	GovKeeper             govkeeper.Keeper
	CrisisKeeper          *crisiskeeper.Keeper
	UpgradeKeeper         *upgradekeeper.Keeper
	NftKeeper             nftkeeper.Keeper
	ParamsKeeper          paramskeeper.Keeper
	IBCKeeper             *ibckeeper.Keeper // IBC Keeper must be a pointer in the app, so we can SetRouter on it correctly
	EvidenceKeeper        evidencekeeper.Keeper
	TransferKeeper        ibctransferkeeper.Keeper
	ICAHostKeeper         icahostkeeper.Keeper
	FeeGrantKeeper        feegrantkeeper.Keeper
	GroupKeeper           groupkeeper.Keeper
	WasmKeeper            wasm.Keeper
	ConsensusParamsKeeper consensuskeeper.Keeper
	AllianceKeeper        alliancekeeper.Keeper

	// bm is the basic module manager
	bm module.BasicManager

	// mm is the module manager
	mm *module.Manager

	// sm is the simulation manager
	sm           *module.SimulationManager
	configurator module.Configurator
}

// New returns a reference to an initialized blockchain app
func New(
	logger log.Logger,
	db tmdb.DB,
	traceStore io.Writer,
	loadLatest bool,
	skipUpgradeHeights map[int64]bool,
	homePath string,
	invCheckPeriod uint,
	encodingConfig appparams.EncodingConfig,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) *App {
	appCodec := encodingConfig.Marshaler
	cdc := encodingConfig.Amino
	interfaceRegistry := encodingConfig.InterfaceRegistry

	// Init App
	app := &App{
		BaseApp: baseapp.NewBaseApp(
			Name,
			logger,
			db,
			encodingConfig.TxConfig.TxDecoder(),
			baseAppOptions...,
		),
		cdc:               cdc,
		appCodec:          appCodec,
		interfaceRegistry: interfaceRegistry,
		invCheckPeriod:    invCheckPeriod,
		keys:              make(map[string]*storetypes.KVStoreKey),
		tkeys:             make(map[string]*storetypes.TransientStoreKey),
		memKeys:           make(map[string]*storetypes.MemoryStoreKey),
	}
	app.SetCommitMultiStoreTracer(traceStore)
	app.SetVersion(version.Version)
	app.SetInterfaceRegistry(interfaceRegistry)

	var basicModules []module.AppModuleBasic = make([]module.AppModuleBasic, 0)
	var modules []module.AppModule = make([]module.AppModule, 0)
	var simModules []module.AppModuleSimulation = make([]module.AppModuleSimulation, 0)

	// 'auth' module
	app.keys[authtypes.StoreKey] = storetypes.NewKVStoreKey(authtypes.StoreKey)
	app.AuthKeeper = authkeeper.NewAccountKeeper(
		appCodec,
		app.keys[authtypes.StoreKey],
		authtypes.ProtoBaseAccount,
		make(map[string][]string), // This will be populated by each module later
		sdktypes.Bech32MainPrefix, // TODO: This might be wrong
		authtypes.NewModuleAddress(govtypes.ModuleName).String(), // TODO: Find out what authority means
	)
	defer func() { // TODO: Does deferring this even work?
		app.AuthKeeper.GetModulePermissions()[authtypes.FeeCollectorName] = authtypes.NewPermissionsForAddress(authtypes.FeeCollectorName, nil) // This implicitly creates a module account
		app.BankKeeper.GetBlockedAddresses()[authtypes.NewModuleAddress(authtypes.FeeCollectorName).String()] = true
	}()
	basicModules = append(basicModules, auth.AppModuleBasic{})
	modules = append(modules, auth.NewAppModule(appCodec, app.AuthKeeper, nil, nil))
	simModules = append(simModules, auth.NewAppModule(appCodec, app.AuthKeeper, authsim.RandomGenesisAccounts, nil)) // TODO: Is RandomGenesisAccounts right?

	// 'bank' module - depends on
	// 1. 'auth'
	app.keys[banktypes.StoreKey] = storetypes.NewKVStoreKey(banktypes.StoreKey)
	app.BankKeeper = alliancebankkeeper.NewBaseKeeper( // Use 'alliance' module's custom implementation instead
		appCodec,
		app.keys[banktypes.StoreKey],
		app.AuthKeeper,
		make(map[string]bool),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	basicModules = append(basicModules, bank.AppModuleBasic{})
	modules = append(modules, alliancebank.NewAppModule(appCodec, app.BankKeeper, app.AuthKeeper, nil))
	simModules = append(simModules, alliancebank.NewAppModule(appCodec, app.BankKeeper, app.AuthKeeper, nil))

	// 'authz' module - depends on
	// 1. 'auth'
	// 2. 'bank'
	app.keys[authzkeeper.StoreKey] = storetypes.NewKVStoreKey(authzkeeper.StoreKey)
	app.AuthzKeeper = authzkeeper.NewKeeper(
		app.keys[authzkeeper.StoreKey],
		appCodec,
		app.MsgServiceRouter(), // TODO: Find out what this is
		app.AuthKeeper,
	)
	basicModules = append(basicModules, authzmodule.AppModuleBasic{})
	modules = append(modules, authzmodule.NewAppModule(appCodec, app.AuthzKeeper, app.AuthKeeper, app.BankKeeper, app.interfaceRegistry))
	simModules = append(simModules, authzmodule.NewAppModule(appCodec, app.AuthzKeeper, app.AuthKeeper, app.BankKeeper, app.interfaceRegistry))

	// 'capability' module
	app.keys[capabilitytypes.StoreKey] = storetypes.NewKVStoreKey(capabilitytypes.StoreKey)
	app.memKeys[capabilitytypes.MemStoreKey] = storetypes.NewMemoryStoreKey(capabilitytypes.MemStoreKey)
	app.CapabilityKeeper = capabilitykeeper.NewKeeper(
		appCodec,
		app.keys[capabilitytypes.StoreKey],
		app.memKeys[capabilitytypes.MemStoreKey],
	)
	defer app.CapabilityKeeper.Seal()
	basicModules = append(basicModules, capability.AppModuleBasic{})
	modules = append(modules, capability.NewAppModule(appCodec, *app.CapabilityKeeper, false)) // TODO: Find out what is sealkeeper
	simModules = append(simModules, capability.NewAppModule(appCodec, *app.CapabilityKeeper, false))

	// 'consensus' module
	app.keys[consensustypes.StoreKey] = storetypes.NewKVStoreKey(consensustypes.StoreKey)
	app.ConsensusParamsKeeper = consensuskeeper.NewKeeper(
		appCodec,
		app.keys[consensustypes.StoreKey],
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	app.SetParamStore(&app.ConsensusParamsKeeper)
	basicModules = append(basicModules, consensus.AppModuleBasic{})
	modules = append(modules, consensus.NewAppModule(appCodec, app.ConsensusParamsKeeper))

	// 'crisis' module - depends on
	// 1. 'bank'
	app.keys[crisistypes.StoreKey] = storetypes.NewKVStoreKey(crisistypes.StoreKey)
	app.CrisisKeeper = crisiskeeper.NewKeeper(
		appCodec,
		app.keys[crisistypes.StoreKey],
		invCheckPeriod, // TODO: Find out what this is
		app.BankKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	basicModules = append(basicModules, crisis.AppModuleBasic{})
	modules = append(modules, crisis.NewAppModule(app.CrisisKeeper, cast.ToBool(appOpts.Get(crisis.FlagSkipGenesisInvariants)), nil))

	// 'feegrant' module - depends on
	// 1. 'auth'
	// 2. 'bank'
	app.keys[feegrant.StoreKey] = storetypes.NewKVStoreKey(feegrant.StoreKey)
	app.FeeGrantKeeper = feegrantkeeper.NewKeeper(
		appCodec,
		app.keys[feegrant.StoreKey],
		app.AuthKeeper,
	)
	basicModules = append(basicModules, feegrantmodule.AppModuleBasic{})
	modules = append(modules, feegrantmodule.NewAppModule(appCodec, app.AuthKeeper, app.BankKeeper, app.FeeGrantKeeper, app.interfaceRegistry))
	simModules = append(simModules, feegrantmodule.NewAppModule(appCodec, app.AuthKeeper, app.BankKeeper, app.FeeGrantKeeper, app.interfaceRegistry))

	// 'group' module - depends on
	// 1. 'auth'
	// 2. 'bank'
	app.keys[group.StoreKey] = storetypes.NewKVStoreKey(group.StoreKey)
	app.GroupKeeper = groupkeeper.NewKeeper(
		app.keys[group.StoreKey],
		appCodec,
		app.MsgServiceRouter(),
		app.AuthKeeper,
		group.DefaultConfig(),
	)
	basicModules = append(basicModules, groupmodule.AppModuleBasic{})
	modules = append(modules, groupmodule.NewAppModule(appCodec, app.GroupKeeper, app.AuthKeeper, app.BankKeeper, app.interfaceRegistry))
	simModules = append(simModules, groupmodule.NewAppModule(appCodec, app.GroupKeeper, app.AuthKeeper, app.BankKeeper, app.interfaceRegistry))

	// 'staking' module - depends on
	// 1. 'auth'
	// 2. 'bank'
	app.AuthKeeper.GetModulePermissions()[stakingtypes.BondedPoolName] = authtypes.NewPermissionsForAddress(stakingtypes.BondedPoolName, []string{authtypes.Burner, authtypes.Staking})
	app.BankKeeper.GetBlockedAddresses()[authtypes.NewModuleAddress(stakingtypes.BondedPoolName).String()] = true
	app.AuthKeeper.GetModulePermissions()[stakingtypes.NotBondedPoolName] = authtypes.NewPermissionsForAddress(stakingtypes.NotBondedPoolName, []string{authtypes.Burner, authtypes.Staking})
	app.BankKeeper.GetBlockedAddresses()[authtypes.NewModuleAddress(stakingtypes.NotBondedPoolName).String()] = true
	app.keys[stakingtypes.StoreKey] = storetypes.NewKVStoreKey(stakingtypes.StoreKey)
	app.StakingKeeper = stakingkeeper.NewKeeper(
		appCodec,
		app.keys[stakingtypes.StoreKey],
		app.AuthKeeper,
		app.BankKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	basicModules = append(basicModules, staking.AppModuleBasic{})
	modules = append(modules, staking.NewAppModule(appCodec, app.StakingKeeper, app.AuthKeeper, app.BankKeeper, nil))
	simModules = append(simModules, staking.NewAppModule(appCodec, app.StakingKeeper, app.AuthKeeper, app.BankKeeper, nil))

	// 'mint' module - depends on
	// 1. 'staking'
	// 2. 'auth'
	// 3. 'bank'
	app.AuthKeeper.GetModulePermissions()[minttypes.ModuleName] = authtypes.NewPermissionsForAddress(minttypes.ModuleName, []string{authtypes.Minter})
	app.BankKeeper.GetBlockedAddresses()[authtypes.NewModuleAddress(minttypes.ModuleName).String()] = true
	app.keys[minttypes.StoreKey] = storetypes.NewKVStoreKey(minttypes.StoreKey)
	app.MintKeeper = mintkeeper.NewKeeper(
		appCodec,
		app.keys[minttypes.StoreKey],
		app.StakingKeeper,
		app.AuthKeeper,
		app.BankKeeper,
		authtypes.FeeCollectorName, // TODO: Find out what this is
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	basicModules = append(basicModules, mint.AppModuleBasic{})
	modules = append(modules, mint.NewAppModule(appCodec, app.MintKeeper, app.AuthKeeper, nil, nil))
	simModules = append(simModules, mint.NewAppModule(appCodec, app.MintKeeper, app.AuthKeeper, nil, nil))

	// 'nft' module - depends on
	// 1. 'auth'
	// 2. 'bank'
	app.AuthKeeper.GetModulePermissions()[nft.ModuleName] = authtypes.NewPermissionsForAddress(nft.ModuleName, nil)
	app.BankKeeper.GetBlockedAddresses()[authtypes.NewModuleAddress(nft.ModuleName).String()] = true
	app.keys[nftkeeper.StoreKey] = storetypes.NewKVStoreKey(nftkeeper.StoreKey)
	app.NftKeeper = nftkeeper.NewKeeper(app.keys[nftkeeper.StoreKey], appCodec, app.AuthKeeper, app.BankKeeper)
	basicModules = append(basicModules, nftmodule.AppModuleBasic{})
	modules = append(modules, nftmodule.NewAppModule(appCodec, app.NftKeeper, app.AuthKeeper, app.BankKeeper, app.interfaceRegistry))
	simModules = append(simModules, nftmodule.NewAppModule(appCodec, app.NftKeeper, app.AuthKeeper, app.BankKeeper, app.interfaceRegistry))

	// 'slashing' module - depends on
	// 1. 'staking'
	// 2. 'auth'
	// 3. 'bank'
	app.keys[slashingtypes.StoreKey] = storetypes.NewKVStoreKey(slashingtypes.StoreKey)
	app.SlashingKeeper = slashingkeeper.NewKeeper(
		appCodec,
		encodingConfig.Amino,
		app.keys[slashingtypes.StoreKey],
		app.StakingKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	basicModules = append(basicModules, slashing.AppModuleBasic{})
	modules = append(modules, slashing.NewAppModule(appCodec, app.SlashingKeeper, app.AuthKeeper, app.BankKeeper, app.StakingKeeper, nil))
	simModules = append(simModules, slashing.NewAppModule(appCodec, app.SlashingKeeper, app.AuthKeeper, app.BankKeeper, app.StakingKeeper, nil))

	// 'gov' module - depends on
	// 1. 'auth'
	// 2. 'bank'
	// 3. 'staking'
	app.AuthKeeper.GetModulePermissions()[govtypes.ModuleName] = authtypes.NewPermissionsForAddress(govtypes.ModuleName, []string{authtypes.Burner})
	app.keys[govtypes.StoreKey] = storetypes.NewKVStoreKey(govtypes.StoreKey)
	app.GovKeeper = *govkeeper.NewKeeper(
		appCodec,
		app.keys[govtypes.StoreKey],
		app.AuthKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		app.MsgServiceRouter(),
		govtypes.DefaultConfig(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	// Set legacy router for backwards compatibility with gov v1beta1
	govLegacyRouter := govv1beta1.NewRouter()
	defer app.GovKeeper.SetLegacyRouter(govLegacyRouter)
	govLegacyRouter.AddRoute(govtypes.RouterKey, govv1beta1.ProposalHandler)
	basicModules = append(basicModules, gov.NewAppModuleBasic(getGovProposalHandlers())) // TODO: Do we need the legacy proposal handlers?
	modules = append(modules, gov.NewAppModule(appCodec, &app.GovKeeper, app.AuthKeeper, app.BankKeeper, nil))
	simModules = append(simModules, gov.NewAppModule(appCodec, &app.GovKeeper, app.AuthKeeper, app.BankKeeper, nil))

	// 'distribution' module - depends on
	// 1. 'auth'
	// 2. 'bank'
	// 3. 'staking'
	// 4. 'gov'
	app.AuthKeeper.GetModulePermissions()[distrtypes.ModuleName] = authtypes.NewPermissionsForAddress(distrtypes.ModuleName, nil)
	app.BankKeeper.GetBlockedAddresses()[authtypes.NewModuleAddress(distrtypes.ModuleName).String()] = true
	app.keys[distrtypes.StoreKey] = storetypes.NewKVStoreKey(distrtypes.StoreKey)
	app.DistrKeeper = distrkeeper.NewKeeper(
		appCodec,
		app.keys[distrtypes.StoreKey],
		app.AuthKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	basicModules = append(basicModules, distr.AppModuleBasic{})
	modules = append(modules, distr.NewAppModule(appCodec, app.DistrKeeper, app.AuthKeeper, app.BankKeeper, app.StakingKeeper, nil))
	simModules = append(simModules, distr.NewAppModule(appCodec, app.DistrKeeper, app.AuthKeeper, app.BankKeeper, app.StakingKeeper, nil))

	// 'params' module - depends on
	// 1. 'gov'
	app.keys[paramstypes.StoreKey] = storetypes.NewKVStoreKey(paramstypes.StoreKey)
	app.tkeys[paramstypes.TStoreKey] = storetypes.NewTransientStoreKey(paramstypes.TStoreKey)
	app.ParamsKeeper = paramskeeper.NewKeeper(
		appCodec,
		cdc,
		app.keys[paramstypes.StoreKey],
		app.tkeys[paramstypes.TStoreKey],
	)
	govLegacyRouter.AddRoute(paramsproposaltypes.RouterKey, params.NewParamChangeProposalHandler(app.ParamsKeeper))
	basicModules = append(basicModules, params.AppModuleBasic{})
	modules = append(modules, params.NewAppModule(app.ParamsKeeper))
	simModules = append(simModules, params.NewAppModule(app.ParamsKeeper))

	// 'evidence' module - depends on
	// 1. 'staking'
	// 2. 'slashing'
	app.keys[evidencetypes.StoreKey] = storetypes.NewKVStoreKey(evidencetypes.StoreKey)
	app.EvidenceKeeper = *evidencekeeper.NewKeeper(
		appCodec,
		app.keys[evidencetypes.StoreKey],
		app.StakingKeeper,
		app.SlashingKeeper,
	)
	basicModules = append(basicModules, evidence.AppModuleBasic{})
	modules = append(modules, evidence.NewAppModule(app.EvidenceKeeper))
	simModules = append(simModules, evidence.NewAppModule(app.EvidenceKeeper))

	// 'upgrade' module - depends on
	// 1. 'gov'
	app.keys[upgradetypes.StoreKey] = storetypes.NewKVStoreKey(upgradetypes.StoreKey)
	app.UpgradeKeeper = upgradekeeper.NewKeeper(
		skipUpgradeHeights, // TODO: What is this?
		app.keys[upgradetypes.StoreKey],
		appCodec,
		homePath,
		app.BaseApp, // TODO: Maybe pass app?
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	govLegacyRouter.AddRoute(upgradetypes.RouterKey, upgrade.NewSoftwareUpgradeProposalHandler(app.UpgradeKeeper))
	basicModules = append(basicModules, upgrade.AppModuleBasic{})
	modules = append(modules, upgrade.NewAppModule(app.UpgradeKeeper))

	// 'ibc' module - depends on
	// 1. 'staking'
	// 2. 'upgrade'
	// 3. 'capability'
	// 4. 'gov'
	// 5. 'params'
	app.keys[ibcexported.StoreKey] = storetypes.NewKVStoreKey(ibcexported.StoreKey)
	app.IBCKeeper = ibckeeper.NewKeeper(
		appCodec,
		app.keys[ibcexported.StoreKey],
		app.ParamsKeeper.Subspace(ibcexported.ModuleName),
		app.StakingKeeper,
		app.UpgradeKeeper,
		app.CapabilityKeeper.ScopeToModule(ibcexported.ModuleName),
	)
	// app.IBCKeeper.SetRouter(ibcporttypes.NewRouter())
	govLegacyRouter.AddRoute(ibcexported.RouterKey, ibcclient.NewClientProposalHandler(app.IBCKeeper.ClientKeeper))
	basicModules = append(basicModules, ibc.AppModuleBasic{})
	modules = append(modules, ibc.NewAppModule(app.IBCKeeper))
	simModules = append(simModules, ibc.NewAppModule(app.IBCKeeper))

	// 'ibctransfer' module - depends on
	// 1. 'ibc'
	// 2. 'auth'
	// 3. 'bank'
	// 4. 'capability'
	app.AuthKeeper.GetModulePermissions()[ibctransfertypes.ModuleName] = authtypes.NewPermissionsForAddress(ibctransfertypes.ModuleName, []string{authtypes.Minter, authtypes.Burner})
	app.BankKeeper.GetBlockedAddresses()[authtypes.NewModuleAddress(ibctransfertypes.ModuleName).String()] = true
	app.keys[ibctransfertypes.StoreKey] = storetypes.NewKVStoreKey(ibctransfertypes.StoreKey)
	app.TransferKeeper = ibctransferkeeper.NewKeeper(
		appCodec,
		app.keys[ibctransfertypes.StoreKey],
		app.ParamsKeeper.Subspace(ibctransfertypes.ModuleName),
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ChannelKeeper,
		&app.IBCKeeper.PortKeeper,
		app.AuthKeeper,
		app.BankKeeper,
		app.CapabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName),
	)
	// app.IBCKeeper.Router.AddRoute(ibctransfertypes.ModuleName, transfer.NewIBCModule(app.TransferKeeper))
	basicModules = append(basicModules, ibctransfer.AppModuleBasic{})
	modules = append(modules, ibctransfer.NewAppModule(app.TransferKeeper))
	simModules = append(simModules, ibctransfer.NewAppModule(app.TransferKeeper))

	// 'ica'
	app.AuthKeeper.GetModulePermissions()[icatypes.ModuleName] = authtypes.NewPermissionsForAddress(icatypes.ModuleName, nil)
	app.BankKeeper.GetBlockedAddresses()[authtypes.NewModuleAddress(icatypes.ModuleName).String()] = true

	// 'icacontroller' module - depends on
	// 1. 'ibc'
	// 2. 'capability'
	app.keys[icacontrollertypes.StoreKey] = storetypes.NewKVStoreKey(icacontrollertypes.StoreKey)
	icaControllerKeeper := icacontrollerkeeper.NewKeeper(
		appCodec,
		app.keys[icacontrollertypes.StoreKey],
		app.ParamsKeeper.Subspace(icacontrollertypes.SubModuleName),
		app.IBCKeeper.ChannelKeeper, // may be replaced with middleware such as ics29 fee
		app.IBCKeeper.ChannelKeeper,
		&app.IBCKeeper.PortKeeper,
		app.CapabilityKeeper.ScopeToModule(icacontrollertypes.SubModuleName),
		app.MsgServiceRouter(),
	)

	// 'icahost' module - depends on
	// 1. 'ibc'
	// 2. 'auth'
	// 3. 'capability'
	// 4. 'icacontroller'
	app.keys[icahosttypes.StoreKey] = storetypes.NewKVStoreKey(icahosttypes.StoreKey)
	app.ICAHostKeeper = icahostkeeper.NewKeeper(
		appCodec,
		app.keys[icahosttypes.StoreKey],
		app.ParamsKeeper.Subspace(icahosttypes.SubModuleName),
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ChannelKeeper,
		&app.IBCKeeper.PortKeeper,
		app.AuthKeeper,
		app.CapabilityKeeper.ScopeToModule(icahosttypes.SubModuleName),
		app.MsgServiceRouter(),
	)
	// app.IBCKeeper.Router.AddRoute(icahosttypes.SubModuleName, icahost.NewIBCModule(app.ICAHostKeeper))
	basicModules = append(basicModules, ica.AppModuleBasic{})
	modules = append(modules, ica.NewAppModule(&icaControllerKeeper, &app.ICAHostKeeper))
	simModules = append(simModules, ica.NewAppModule(&icaControllerKeeper, &app.ICAHostKeeper))

	// 'wasm' module - depends on
	// 1. 'gov'
	// 2. 'auth'
	// 3. 'bank'
	// 4. 'staking'
	// 5. 'distribution'
	// 6. 'capability'
	// 7. 'ibc'
	// 8. 'ibctransfer'
	app.AuthKeeper.GetModulePermissions()[wasmtypes.ModuleName] = authtypes.NewPermissionsForAddress(wasmtypes.ModuleName, []string{authtypes.Burner})
	app.BankKeeper.GetBlockedAddresses()[authtypes.NewModuleAddress(wasmtypes.ModuleName).String()] = true
	app.keys[wasmtypes.StoreKey] = storetypes.NewKVStoreKey(wasmtypes.StoreKey)
	wasmConfig, err := wasm.ReadWasmConfig(appOpts)
	if err != nil {
		panic(fmt.Sprintf("error while reading wasm config: %s", err))
	}
	app.WasmKeeper = wasm.NewKeeper(
		appCodec,
		app.keys[wasm.StoreKey],
		app.AuthKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		distrkeeper.NewQuerier(app.DistrKeeper),
		app.IBCKeeper.ChannelKeeper,
		&app.IBCKeeper.PortKeeper,
		app.CapabilityKeeper.ScopeToModule(wasm.ModuleName),
		app.TransferKeeper,
		app.MsgServiceRouter(),
		app.GRPCQueryRouter(),
		filepath.Join(homePath, "wasm"),
		wasmConfig,
		"iterator,staking,stargate,cosmwasm_1_1", // TODO: Find out what this configures
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	govLegacyRouter.AddRoute(wasm.RouterKey, wasm.NewWasmProposalHandler(app.WasmKeeper, wasm.EnableAllProposals))
	basicModules = append(basicModules, wasm.AppModuleBasic{})
	modules = append(modules, wasm.NewAppModule(appCodec, &app.WasmKeeper, app.StakingKeeper, app.AuthKeeper, app.BankKeeper, app.MsgServiceRouter(), nil))
	simModules = append(simModules, wasm.NewAppModule(appCodec, &app.WasmKeeper, app.StakingKeeper, app.AuthKeeper, app.BankKeeper, app.MsgServiceRouter(), nil))

	// 'alliance' module - depends on
	// 1. 'auth'
	// 2. 'bank'
	// 3. 'staking'
	// 4. 'distribution'
	// 5. 'gov'
	app.BankKeeper.RegisterKeepers(app.AllianceKeeper, app.StakingKeeper)
	app.AuthKeeper.GetModulePermissions()[alliancetypes.ModuleName] = authtypes.NewPermissionsForAddress(alliancetypes.ModuleName, []string{authtypes.Minter, authtypes.Burner})
	app.AuthKeeper.GetModulePermissions()[alliancetypes.RewardsPoolName] = authtypes.NewPermissionsForAddress(alliancetypes.RewardsPoolName, nil)
	app.BankKeeper.GetBlockedAddresses()[authtypes.NewModuleAddress(alliancetypes.RewardsPoolName).String()] = true
	app.keys[alliancetypes.StoreKey] = storetypes.NewKVStoreKey(alliancetypes.StoreKey)
	app.AllianceKeeper = alliancekeeper.NewKeeper(
		appCodec,
		app.keys[alliancetypes.StoreKey],
		app.ParamsKeeper.Subspace(alliancetypes.ModuleName),
		app.AuthKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		app.DistrKeeper,
	)
	govLegacyRouter.AddRoute(alliancetypes.RouterKey, alliance.NewAllianceProposalHandler(app.AllianceKeeper))
	basicModules = append(basicModules, alliance.AppModuleBasic{})
	// modules = append(modules, alliance.NewAppModule(appCodec, app.AllianceKeeper, app.StakingKeeper, app.AuthKeeper, app.BankKeeper, app.interfaceRegistry))
	// simModules = append(simModules, alliance.NewAppModule(appCodec, app.AllianceKeeper, app.StakingKeeper, app.AuthKeeper, app.BankKeeper, app.interfaceRegistry))

	// 'genutil' module - depends on
	// 1. 'auth'
	// 2. 'staking'
	basicModules = append(basicModules, genutil.AppModuleBasic{})
	modules = append(modules, genutil.NewAppModule(app.AuthKeeper, app.StakingKeeper, app.BaseApp.DeliverTx, encodingConfig.TxConfig))

	// 'vesting' module - depends on
	// 1. 'auth'
	// 2. 'bank'
	basicModules = append(basicModules, vesting.AppModuleBasic{})
	modules = append(modules, vesting.NewAppModule(app.AuthKeeper, app.BankKeeper))

	// post handling
	app.StakingKeeper.SetHooks(stakingtypes.NewMultiStakingHooks(
		app.DistrKeeper.Hooks(),
		app.SlashingKeeper.Hooks(),
		// app.AllianceKeeper.StakingHooks(),
	))

	/****  Module Options ****/

	// NOTE: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.

	app.bm = module.NewBasicManager(basicModules...)
	app.bm.RegisterLegacyAminoCodec(encodingConfig.Amino)
	app.bm.RegisterInterfaces(encodingConfig.InterfaceRegistry)

	app.mm = module.NewManager(modules...)

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	// NOTE: staking module is required if HistoricalEntries param > 0
	app.mm.SetOrderBeginBlockers(
		// upgrades should be run first
		upgradetypes.ModuleName,
		capabilitytypes.ModuleName,
		minttypes.ModuleName,
		consensustypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		evidencetypes.ModuleName,
		stakingtypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		govtypes.ModuleName,
		crisistypes.ModuleName,
		ibctransfertypes.ModuleName,
		ibcexported.ModuleName,
		icatypes.ModuleName,
		genutiltypes.ModuleName,
		authz.ModuleName,
		feegrant.ModuleName,
		group.ModuleName,
		paramstypes.ModuleName,
		vestingtypes.ModuleName,
		nft.ModuleName,
		// alliancetypes.ModuleName,
		wasm.ModuleName,
		// this line is used by starport scaffolding # stargate/app/beginBlockers
	)

	app.mm.SetOrderEndBlockers(
		crisistypes.ModuleName,
		govtypes.ModuleName,
		stakingtypes.ModuleName,
		consensustypes.ModuleName,
		ibctransfertypes.ModuleName,
		ibcexported.ModuleName,
		icatypes.ModuleName,
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		minttypes.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		authz.ModuleName,
		feegrant.ModuleName,
		group.ModuleName,
		paramstypes.ModuleName,
		upgradetypes.ModuleName,
		vestingtypes.ModuleName,
		nft.ModuleName,
		// alliancetypes.ModuleName,
		wasm.ModuleName,
		// this line is used by starport scaffolding # stargate/app/endBlockers
	)

	// NOTE: The genutils module must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	// NOTE: Capability module must occur first so that it can initialize any capabilities
	// so that other modules that want to create or claim capabilities afterwards in InitChain
	// can do so safely.
	app.mm.SetOrderInitGenesis(
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		stakingtypes.ModuleName,
		slashingtypes.ModuleName,
		consensustypes.ModuleName,
		govtypes.ModuleName,
		minttypes.ModuleName,
		crisistypes.ModuleName,
		genutiltypes.ModuleName,
		ibctransfertypes.ModuleName,
		ibcexported.ModuleName,
		icatypes.ModuleName,
		evidencetypes.ModuleName,
		authz.ModuleName,
		feegrant.ModuleName,
		group.ModuleName,
		paramstypes.ModuleName,
		upgradetypes.ModuleName,
		vestingtypes.ModuleName,
		nft.ModuleName,
		// alliancetypes.ModuleName,
		wasm.ModuleName,
		// this line is used by starport scaffolding # stargate/app/initGenesis
	)

	// Uncomment if you want to set a custom migration order here.
	// app.mm.SetOrderMigrations(custom order)

	app.mm.RegisterInvariants(app.CrisisKeeper)

	app.configurator = module.NewConfigurator(app.appCodec, app.MsgServiceRouter(), app.GRPCQueryRouter())
	app.mm.RegisterServices(app.configurator)

	// create the simulation manager and define the order of the modules for deterministic simulations
	app.sm = module.NewSimulationManager(simModules...)
	app.sm.RegisterStoreDecoders()

	// initialize stores
	app.MountKVStores(app.keys)
	app.MountTransientStores(app.tkeys)
	app.MountMemoryStores(app.memKeys)

	// initialize BaseApp
	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)

	anteHandler, err := NewAnteHandler(
		HandlerOptions{
			HandlerOptions: ante.HandlerOptions{
				AccountKeeper:   app.AuthKeeper,
				BankKeeper:      app.BankKeeper,
				SignModeHandler: encodingConfig.TxConfig.SignModeHandler(),
				FeegrantKeeper:  app.FeeGrantKeeper,
				SigGasConsumer:  ante.DefaultSigVerificationGasConsumer,
			},
			IBCKeeper:         app.IBCKeeper,
			TxCounterStoreKey: app.GetKey(wasm.StoreKey),
			WasmConfig:        wasmConfig,
			Cdc:               appCodec,
		},
	)
	if err != nil {
		panic(fmt.Errorf("failed to create AnteHandler: %s", err))
	}

	app.SetAnteHandler(anteHandler)
	app.SetEndBlocker(app.EndBlocker)

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			tmos.Exit(err.Error())
		}
	}

	return app
}

// Name returns the name of the App
func (app *App) Name() string { return app.BaseApp.Name() }

// BeginBlocker application updates every begin block
func (app *App) BeginBlocker(ctx sdktypes.Context, req abcitypes.RequestBeginBlock) abcitypes.ResponseBeginBlock {
	return app.mm.BeginBlock(ctx, req)
}

// EndBlocker application updates every end block
func (app *App) EndBlocker(ctx sdktypes.Context, req abcitypes.RequestEndBlock) abcitypes.ResponseEndBlock {
	return app.mm.EndBlock(ctx, req)
}

// InitChainer application update at chain initialization
func (app *App) InitChainer(ctx sdktypes.Context, req abcitypes.RequestInitChain) abcitypes.ResponseInitChain {
	var genesisState GenesisState
	if err := tmjson.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}
	app.UpgradeKeeper.SetModuleVersionMap(ctx, app.mm.GetVersionMap())
	return app.mm.InitGenesis(ctx, app.appCodec, genesisState)
}

// LoadHeight loads a particular height
func (app *App) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// LegacyAmino returns SimApp's amino codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *App) LegacyAmino() *codec.LegacyAmino {
	return app.cdc
}

// AppCodec returns an app codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *App) AppCodec() codec.Codec {
	return app.appCodec
}

// InterfaceRegistry returns an InterfaceRegistry
func (app *App) InterfaceRegistry() types.InterfaceRegistry {
	return app.interfaceRegistry
}

// GetKey returns the KVStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *App) GetKey(storeKey string) *storetypes.KVStoreKey {
	return app.keys[storeKey]
}

// GetTKey returns the TransientStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *App) GetTKey(storeKey string) *storetypes.TransientStoreKey {
	return app.tkeys[storeKey]
}

// GetMemKey returns the MemStoreKey for the provided mem key.
//
// NOTE: This is solely used for testing purposes.
func (app *App) GetMemKey(storeKey string) *storetypes.MemoryStoreKey {
	return app.memKeys[storeKey]
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *App) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx
	// Register new tx routes from grpc-gateway.
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	// Register new tendermint queries routes from grpc-gateway.
	tmservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register grpc-gateway routes for all modules.
	app.bm.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// register app's OpenAPI routes.
	apiSvr.Router.Handle("/static/openapi.yml", http.FileServer(http.FS(docs.Docs)))
	apiSvr.Router.HandleFunc("/", openapiconsole.Handler(Name, "/static/openapi.yml"))
}

// RegisterTxService implements the Application.RegisterTxService method.
func (app *App) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.BaseApp.Simulate, app.interfaceRegistry)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *App) RegisterTendermintService(clientCtx client.Context) {
	tmservice.RegisterTendermintService(
		clientCtx,
		app.BaseApp.GRPCQueryRouter(),
		app.interfaceRegistry,
		app.Query,
	)
}

func (app *App) RegisterNodeService(clientCtx client.Context) {
	node.RegisterNodeService(clientCtx, app.GRPCQueryRouter())
}

// SimulationManager implements the SimulationApp interface
func (app *App) SimulationManager() *module.SimulationManager {
	return app.sm
}

// DefaultGenesis returns a default genesis from the registered AppModuleBasic's.
func (app *App) DefaultGenesis() map[string]json.RawMessage {
	return app.bm.DefaultGenesis(app.appCodec)
}

func (app *App) Basic() module.BasicManager {
	return app.bm
}
