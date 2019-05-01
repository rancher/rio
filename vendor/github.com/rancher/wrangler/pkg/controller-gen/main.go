package controllergen

import (
	"path/filepath"

	cgargs "github.com/rancher/wrangler/pkg/controller-gen/args"
	"github.com/rancher/wrangler/pkg/controller-gen/generators"
	"k8s.io/apimachinery/pkg/runtime/schema"
	csargs "k8s.io/code-generator/cmd/client-gen/args"
	clientgenerators "k8s.io/code-generator/cmd/client-gen/generators"
	cs "k8s.io/code-generator/cmd/client-gen/generators"
	types2 "k8s.io/code-generator/cmd/client-gen/types"
	dpargs "k8s.io/code-generator/cmd/deepcopy-gen/args"
	infargs "k8s.io/code-generator/cmd/informer-gen/args"
	inf "k8s.io/code-generator/cmd/informer-gen/generators"
	lsargs "k8s.io/code-generator/cmd/lister-gen/args"
	ls "k8s.io/code-generator/cmd/lister-gen/generators"
	"k8s.io/gengo/args"
	dp "k8s.io/gengo/examples/deepcopy-gen/generators"
	"k8s.io/gengo/types"
	"k8s.io/klog"
)

func Run(opts cgargs.Options) {
	customArgs := &cgargs.CustomArgs{
		Options:      opts,
		TypesByGroup: map[schema.GroupVersion][]*types.Name{},
		Package:      opts.OutputPackage,
	}

	genericArgs := args.Default().WithoutDefaultFlagParsing()
	genericArgs.CustomArgs = customArgs
	genericArgs.GoHeaderFilePath = opts.Boilerplate
	genericArgs.InputDirs = parseTypes(customArgs)

	clientGen := generators.NewClientGenerator()

	if err := genericArgs.Execute(
		clientgenerators.NameSystems(),
		clientgenerators.DefaultNameSystem(),
		clientGen.Packages,
	); err != nil {
		klog.Fatalf("Error: %v", err)
	}

	groups := map[string]bool{}
	for groupName, group := range customArgs.Options.Groups {
		if group.GenerateTypes {
			groups[groupName] = true
		}
	}

	if len(groups) == 0 {
		return
	}

	if err := generateDeepcopy(groups, customArgs); err != nil {
		klog.Fatalf("deepcopy failed: %v", err)
	}

	if err := generateClientset(groups, customArgs); err != nil {
		klog.Fatalf("clientset failed: %v", err)
	}

	if err := generateListers(groups, customArgs); err != nil {
		klog.Fatalf("listers failed: %v", err)
	}

	if err := generateInformers(groups, customArgs); err != nil {
		klog.Fatalf("informers failed: %v", err)
	}
	//if err := clientGen.GenerateMocks(); err != nil {
	//	klog.Fatalf("mocks failed: %v", err)
	//}
}

func generateDeepcopy(groups map[string]bool, customArgs *cgargs.CustomArgs) error {
	deepCopyCustomArgs := &dpargs.CustomArgs{}

	args := args.Default().WithoutDefaultFlagParsing()
	args.CustomArgs = deepCopyCustomArgs
	args.OutputFileBaseName = "zz_generated_deepcopy"
	args.GoHeaderFilePath = customArgs.Options.Boilerplate

	for gv, names := range customArgs.TypesByGroup {
		if !groups[gv.Group] {
			continue
		}
		args.InputDirs = append(args.InputDirs, names[0].Package)
		deepCopyCustomArgs.BoundingDirs = append(deepCopyCustomArgs.BoundingDirs, names[0].Package)
	}

	return args.Execute(dp.NameSystems(),
		dp.DefaultNameSystem(),
		dp.Packages)

}

func generateClientset(groups map[string]bool, customArgs *cgargs.CustomArgs) error {
	args, clientSetArgs := csargs.NewDefaults()
	clientSetArgs.ClientsetName = "versioned"
	args.OutputPackagePath = filepath.Join(customArgs.Package, "clientset")
	args.GoHeaderFilePath = customArgs.Options.Boilerplate

	for gv, names := range customArgs.TypesByGroup {
		if !groups[gv.Group] {
			continue
		}
		args.InputDirs = append(args.InputDirs, names[0].Package)
		clientSetArgs.Groups = append(clientSetArgs.Groups, types2.GroupVersions{
			PackageName: gv.Group,
			Group:       types2.Group(gv.Group),
			Versions: []types2.PackageVersion{
				{
					Version: types2.Version(gv.Version),
					Package: names[0].Package,
				},
			},
		})
	}

	return args.Execute(cs.NameSystems(),
		cs.DefaultNameSystem(),
		cs.Packages)
}

func generateInformers(groups map[string]bool, customArgs *cgargs.CustomArgs) error {
	args, clientSetArgs := infargs.NewDefaults()
	clientSetArgs.VersionedClientSetPackage = filepath.Join(customArgs.Package, "clientset/versioned")
	clientSetArgs.ListersPackage = filepath.Join(customArgs.Package, "listers")
	args.OutputPackagePath = filepath.Join(customArgs.Package, "informers")
	args.GoHeaderFilePath = customArgs.Options.Boilerplate

	for gv, names := range customArgs.TypesByGroup {
		if !groups[gv.Group] {
			continue
		}
		args.InputDirs = append(args.InputDirs, names[0].Package)
	}

	return args.Execute(inf.NameSystems(),
		inf.DefaultNameSystem(),
		inf.Packages)
}

func generateListers(groups map[string]bool, customArgs *cgargs.CustomArgs) error {
	args, _ := lsargs.NewDefaults()
	args.OutputPackagePath = filepath.Join(customArgs.Package, "listers")
	args.GoHeaderFilePath = customArgs.Options.Boilerplate

	for gv, names := range customArgs.TypesByGroup {
		if !groups[gv.Group] {
			continue
		}
		args.InputDirs = append(args.InputDirs, names[0].Package)
	}

	return args.Execute(ls.NameSystems(),
		ls.DefaultNameSystem(),
		ls.Packages)
}

func parseTypes(customArgs *cgargs.CustomArgs) []string {
	for groupName, group := range customArgs.Options.Groups {
		if group.GenerateTypes {
			group.InformersPackage = filepath.Join(customArgs.Package, "informers/externalversions")
			group.ClientSetPackage = filepath.Join(customArgs.Package, "clientset/versioned")
			group.ListersPackage = filepath.Join(customArgs.Package, "listers")
			customArgs.Options.Groups[groupName] = group
		}
	}

	for groupName, group := range customArgs.Options.Groups {
		cgargs.ObjectsToGroupVersion(groupName, group.Types, customArgs.TypesByGroup)
	}

	var inputDirs []string
	for _, names := range customArgs.TypesByGroup {
		inputDirs = append(inputDirs, names[0].Package)
	}

	return inputDirs
}
