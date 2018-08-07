// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2018 Giuseppe Maxia
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"github.com/grooverdan/dbdeployer/common"
	"github.com/grooverdan/dbdeployer/sandbox"
	"github.com/spf13/cobra"
	"os"
)

func UnpreserveSandbox(sandbox_dir, sandbox_name string) {
	full_path := sandbox_dir + "/" + sandbox_name
	if !common.DirExists(full_path) {
		common.Exit(1, fmt.Sprintf("Directory '%s' not found", full_path))
	}
	preserve := full_path + "/no_clear_all"
	if !common.ExecExists(preserve) {
		preserve = full_path + "/no_clear"
	}
	if !common.ExecExists(preserve) {
		fmt.Printf("Sandbox %s is not locked\n",sandbox_name)
		return	
	}
	is_multiple := true
	clear := full_path + "/clear_all"
	if !common.ExecExists(clear) {
		clear = full_path + "/clear"
		is_multiple = false
	}
	if !common.ExecExists(clear) {
		common.Exit(1, fmt.Sprintf("Executable '%s' not found", clear))
	}
	no_clear := full_path + "/no_clear"
	if is_multiple {
		no_clear = full_path + "/no_clear_all"
	}
	err := os.Remove(clear)
	if err != nil {
		common.Exit(1, fmt.Sprintf("Error while removing %s \n%s",clear, err))
	}
	err = os.Rename(no_clear, clear)
	if err != nil {
		common.Exit(1, fmt.Sprintf("Error while renaming  script\n%s", err))
	}
	fmt.Printf("Sandbox %s unlocked\n",sandbox_name)
}



func PreserveSandbox(sandbox_dir, sandbox_name string) {
	full_path := sandbox_dir + "/" + sandbox_name
	if !common.DirExists(full_path) {
		common.Exit(1, fmt.Sprintf("Directory '%s' not found", full_path))
	}
	preserve := full_path + "/no_clear_all"
	if !common.ExecExists(preserve) {
		preserve = full_path + "/no_clear"
	}
	if common.ExecExists(preserve) {
		fmt.Printf("Sandbox %s is already locked\n",sandbox_name)
		return	
	}
	is_multiple := true
	clear := full_path + "/clear_all"
	if !common.ExecExists(clear) {
		clear = full_path + "/clear"
		is_multiple = false
	}
	if !common.ExecExists(clear) {
		common.Exit(1, fmt.Sprintf("Executable '%s' not found", clear))
	}
	no_clear := full_path + "/no_clear"
	clear_cmd := "clear"
	no_clear_cmd := "no_clear"
	if is_multiple {
		no_clear = full_path + "/no_clear_all"
		clear_cmd = "clear_all"
		no_clear_cmd = "no_clear_all"
	}
	err := os.Rename(clear, no_clear)
	if err != nil {
		common.Exit(1, fmt.Sprintf( "Error while renaming script.\n%s",err))
	}
	template := sandbox.SingleTemplates["sb_locked_template"].Contents
	var data = common.Smap{
		"TemplateName" : "sb_locked_template",
		"SandboxDir" : sandbox_name,
		"AppVersion" : common.VersionDef,
		"Copyright" : sandbox.Copyright,
		"ClearCmd" : clear_cmd,
		"NoClearCmd" : no_clear_cmd,
	}
	template = common.TrimmedLines(template)
	new_clear_message := common.Tprintf(template, data)
	common.WriteString(new_clear_message, clear)
	os.Chmod(clear, 0744)
	fmt.Printf("Sandbox %s locked\n",sandbox_name)
}

func LockSandbox(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		common.Exit(1, 
			"'lock' requires the name of a sandbox (or ALL)",
			"Example: dbdeployer admin lock msb_5_7_21")
	}
	sandbox := args[0]
	sandbox_dir := GetAbsolutePathFromFlag(cmd, "sandbox-home")
	lock_list := []string{sandbox}
	if sandbox == "ALL" || sandbox == "all" {
		lock_list = common.SandboxInfoToFileNames(common.GetInstalledSandboxes(sandbox_dir))
	}
	if len(lock_list) == 0 {
		fmt.Printf("Nothing to lock in %s\n", sandbox_dir)
		return
	}
	for _, sb := range lock_list {
		PreserveSandbox(sandbox_dir, sb)
	}
}

func UnlockSandbox(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		common.Exit(1, 
			"'unlock' requires the name of a sandbox (or ALL)",
			"Example: dbdeployer admin unlock msb_5_7_21")
	}
	sandbox := args[0]
	sandbox_dir := GetAbsolutePathFromFlag(cmd, "sandbox-home")
	lock_list := []string{sandbox}
	if sandbox == "ALL" || sandbox == "all" {
		lock_list = common.SandboxInfoToFileNames(common.GetInstalledSandboxes(sandbox_dir))
	}
	if len(lock_list) == 0 {
		fmt.Printf("Nothing to lock in %s\n", sandbox_dir)
		return
	}
	for _, sb := range lock_list {
		UnpreserveSandbox(sandbox_dir, sb)
	}
}


var (
	adminCmd = &cobra.Command{
		Use:   "admin",
		Short: "sandbox management tasks",
		Aliases: []string{"manage"},
		Long: `Runs commands related to the administration of sandboxes.`,
	}

	adminLockCmd = &cobra.Command{
		Use:     "lock sandbox_name",
		Aliases: []string{"preserve"},
		Short:   "Locks a sandbox, preventing deletion",
		Long: `Prevents deletion for a given sandbox.
Note that the deletion being prevented is only the one occurring through dbdeployer. 
Users can still delete locked sandboxes manually.`,
		Run: LockSandbox,
	}

	adminUnlockCmd = &cobra.Command{
		Use:     "unlock sandbox_name",
		Aliases: []string{"unpreserve"},
		Short:   "Unlocks a sandbox",
		Long: `Removes lock, allowing deletion of a given sandbox`,
		Run: UnlockSandbox,
	}
)

func init() {
	rootCmd.AddCommand(adminCmd)
	adminCmd.AddCommand(adminLockCmd)
	adminCmd.AddCommand(adminUnlockCmd)
}
