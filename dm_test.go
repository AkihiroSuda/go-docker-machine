package dm

import (
	"os/exec"
	"testing"

	"golang.org/x/net/context"
)

func ensureDM(t *testing.T) *DockerMachine {
	_, err := exec.LookPath(DefaultDockerMachinePath)
	if err != nil {
		t.Skipf("%s seems not installed: %v",
			DefaultDockerMachinePath, err)
	}
	return NewDockerMachine()
}

func TestDMVersion(t *testing.T) {
	dm := ensureDM(t)
	res, err := dm.RawCommandResult("version")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("version result: \"%s\"", string(res))
}

func TestMachines(t *testing.T) {
	dm := ensureDM(t)
	machines, err := dm.Machines()
	if err != nil {
		t.Fatal(err)
	}
	for i, machine := range machines {
		t.Logf("machine %d: \"%s\" (State:\"%s\", URL:\"%s\")", i,
			machine.Name, machine.State, machine.URL)
	}
}

func TestInspect(t *testing.T) {
	dm := ensureDM(t)
	machines, err := dm.Machines()
	if err != nil {
		t.Fatal(err)
	}
	for _, machine := range machines {
		name := machine.Name
		config, err := dm.Inspect(name)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("// config for %s", name)
		t.Logf("%#v", config)
	}
}

func TestClient(t *testing.T) {
	dm := ensureDM(t)
	machines, err := dm.Machines()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	for _, machine := range machines {
		name := machine.Name
		if machine.State != "Running" {
			t.Logf("// %s is not running (%v)", name, machine.State)
			continue
		}
		client, err := dm.Client(name)
		if err != nil {
			t.Fatal(err)
		}
		info, err := client.Info(ctx)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("// info for %s", name)
		t.Logf("%#v", info)
	}
}

func TestEnv(t *testing.T) {
	dm := ensureDM(t)
	machines, err := dm.Machines()
	if err != nil {
		t.Fatal(err)
	}
	for _, machine := range machines {
		name := machine.Name
		if machine.State != "Running" {
			t.Logf("// %s is not running (%v)", name, machine.State)
			continue
		}
		env, err := dm.Env(name)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("// env for %s", name)
		t.Logf("%#v", env)
	}
}
