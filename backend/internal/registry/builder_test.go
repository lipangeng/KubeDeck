package registry

import "testing"

func TestBuildSnapshot(t *testing.T) {
	system := Snapshot{
		ResourceTypes: []ResourceType{
			{
				ID:               "apps-v1-deployments",
				Group:            "apps",
				Version:          "v1",
				Kind:             "Deployment",
				Plural:           "deployments",
				Namespaced:       true,
				PreferredVersion: "v1",
				Source:           "system",
			},
		},
		Pages: []PageMeta{
			{PageID: "workloads", Route: "/workloads", PluginID: "system"},
		},
		Slots: []SlotMeta{
			{SlotID: "workloads.tabs", PageID: "workloads", Accepts: "tab", Ordering: "weight"},
		},
		Menus: []MenuItem{
			{
				ID:           "menu-workloads",
				Group:        "workloads",
				TitleI18nKey: "menu.workloads",
				TargetType:   "page",
				TargetRef:    "workloads",
				Source:       "system",
				Order:        10,
				Visible:      true,
			},
		},
	}

	dynamic := Snapshot{
		ResourceTypes: []ResourceType{
			{
				ID:               "argoproj-io-v1alpha1-applications",
				Group:            "argoproj.io",
				Version:          "v1alpha1",
				Kind:             "Application",
				Plural:           "applications",
				Namespaced:       true,
				PreferredVersion: "v1alpha1",
				Source:           "dynamic-crd",
			},
		},
		Menus: []MenuItem{
			{
				ID:           "menu-argocd-apps",
				Group:        "argoproj.io",
				TitleI18nKey: "menu.argocd.applications",
				TargetType:   "resource",
				TargetRef:    "argoproj-io-v1alpha1-applications",
				Source:       "dynamic",
				Order:        100,
				Visible:      true,
			},
		},
	}

	snapshot := BuildSnapshot(system, dynamic)

	if got, want := len(snapshot.ResourceTypes), 2; got != want {
		t.Fatalf("expected %d resource types, got %d", want, got)
	}
	if got, want := len(snapshot.Pages), 1; got != want {
		t.Fatalf("expected %d pages, got %d", want, got)
	}
	if got, want := len(snapshot.Slots), 1; got != want {
		t.Fatalf("expected %d slots, got %d", want, got)
	}
	if got, want := len(snapshot.Menus), 2; got != want {
		t.Fatalf("expected %d menus, got %d", want, got)
	}

	if snapshot.ResourceTypes[0].Source != "system" {
		t.Fatalf("expected first resource source system, got %q", snapshot.ResourceTypes[0].Source)
	}
	if snapshot.ResourceTypes[1].Source != "dynamic-crd" {
		t.Fatalf("expected second resource source dynamic-crd, got %q", snapshot.ResourceTypes[1].Source)
	}
	if snapshot.Menus[0].Source != "system" {
		t.Fatalf("expected first menu source system, got %q", snapshot.Menus[0].Source)
	}
	if snapshot.Menus[1].Source != "dynamic" {
		t.Fatalf("expected second menu source dynamic, got %q", snapshot.Menus[1].Source)
	}
}

func TestBuildSnapshotDeepCopiesNestedSlices(t *testing.T) {
	system := Snapshot{
		Pages: []PageMeta{
			{
				PageID:   "workloads",
				Route:    "/workloads",
				PluginID: "system",
				Slots:    []string{"overview", "events"},
			},
		},
		Menus: []MenuItem{
			{
				ID:              "menu-workloads",
				PermissionHints: []string{"view.workloads", "view.events"},
			},
		},
	}
	dynamic := Snapshot{
		Pages: []PageMeta{
			{
				PageID:   "crd-apps",
				Route:    "/crd/apps",
				PluginID: "dynamic",
				Slots:    []string{"details"},
			},
		},
		Menus: []MenuItem{
			{
				ID:              "menu-dynamic-apps",
				PermissionHints: []string{"view.dynamic"},
			},
		},
	}

	snapshot := BuildSnapshot(system, dynamic)

	system.Pages[0].Slots[0] = "mutated-system-slot"
	system.Menus[0].PermissionHints[0] = "mutated.system.permission"
	dynamic.Pages[0].Slots[0] = "mutated-dynamic-slot"
	dynamic.Menus[0].PermissionHints[0] = "mutated.dynamic.permission"

	if got := snapshot.Pages[0].Slots[0]; got != "overview" {
		t.Fatalf("expected system slot to stay %q, got %q", "overview", got)
	}
	if got := snapshot.Pages[1].Slots[0]; got != "details" {
		t.Fatalf("expected dynamic slot to stay %q, got %q", "details", got)
	}
	if got := snapshot.Menus[0].PermissionHints[0]; got != "view.workloads" {
		t.Fatalf("expected system permission hint to stay %q, got %q", "view.workloads", got)
	}
	if got := snapshot.Menus[1].PermissionHints[0]; got != "view.dynamic" {
		t.Fatalf("expected dynamic permission hint to stay %q, got %q", "view.dynamic", got)
	}
}
