# Exploration Summary
Enabling support for running .Net Core applications on top of the Tiny stack required only a few, relatively simple changes to the existing buildpacks. I've outlined those changes below, but in summary, we need to 1) update the stack references that are supported across all the `buildpack.toml` files, 2) make some small changes to the internal implementatons to support running on top of a stack that does not have a shell, and 3) include a couple of new packages in the Tiny Run Image that are required at runtime.

I believe the biggest hurdle to supporting .Net Core applications on Tiny is the conversation around the additions to the Tiny stack. I don't know much about the goals of Tiny and whether these additions would be welcome or considered to be outside of the scope of the stack. FWIW, the additions here seem to only amount to an ~2MB increase in the stack image size.

## Buildpack Changes

### dotnet-core-aspnet
1. Add `io.paketo.stacks.tiny` to the `stacks` list in the `buildpack.toml`
1. Include `io.paketo.stacks.tiny` on the list of `stacks` for all `metadata.dependencies` in the `buildpack.toml`
```diff
diff --git a/buildpack.toml b/buildpack.toml
index 90a8df9..8111c2d 100644
--- a/buildpack.toml
+++ b/buildpack.toml
@@ -15,7 +15,7 @@ api = "0.2"
    sha256 = "42caecb083385584978bc246987a4b86f88680ed8d2f950a131d3a27b1562870"
    source = "https://download.visualstudio.microsoft.com/download/pr/1d6ae2ec-4cf8-4579-bdfb-18c723b1a560/48be79a406578690a3f062ff17d663f8/aspnetcore-runtime-2.1.21-linux-x64.tar.gz"
    source_sha256 = "75dc48d0fe0cba6f80cfe017b9c3f57908efd87ffe3243956b59b8bfb421e369"
-    stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.cflinuxfs3"]
+    stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.cflinuxfs3", "io.paketo.stacks.tiny"]
    uri = "https://buildpacks.cloudfoundry.org/dependencies/dotnet-aspnetcore/dotnet-aspnetcore_2.1.21_linux_x64_any-stack_42caecb0.tar.xz"
    version = "2.1.21"

@@ -24,7 +24,7 @@ api = "0.2"
    sha256 = "66e1b0f28c3603ae4ac1f120da0b23f40947e08eb7ed8e898549c1b2f4216a73"
    source = "https://download.visualstudio.microsoft.com/download/pr/c1798274-4f4e-4e5b-8337-cb477add793c/2ab1c7f92fe497e07304b0b25c5f7845/aspnetcore-runtime-2.1.22-linux-x64.tar.gz"
    source_sha256 = "053cb445608296a5c6d988980bdbfe1ee36602d1445fd67835d29eceab916ef0"
-    stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.cflinuxfs3"]
+    stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.cflinuxfs3", "io.paketo.stacks.tiny"]
    uri = "https://buildpacks.cloudfoundry.org/dependencies/dotnet-aspnetcore/dotnet-aspnetcore_2.1.22_linux_x64_any-stack_66e1b0f2.tar.xz"
    version = "2.1.22"

@@ -33,7 +33,7 @@ api = "0.2"
    sha256 = "50ddfdfc8bb28984cdbac100c098bd61828f8706df72e7899d3c2b04c7628db0"
    source = "https://download.visualstudio.microsoft.com/download/pr/e7d0601d-41b4-483f-b411-f2b42708054a/191b56b81e1830b413d0794728831eea/aspnetcore-runtime-3.1.7-linux-x64.tar.gz"
    source_sha256 = "4f0ce619c1b1dbc8ccd799877b5d73158a07b1ebd1222d44b909bba13bdf735c"
-    stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.cflinuxfs3"]
+    stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.cflinuxfs3", "io.paketo.stacks.tiny"]
    uri = "https://buildpacks.cloudfoundry.org/dependencies/dotnet-aspnetcore/dotnet-aspnetcore_3.1.7_linux_x64_any-stack_50ddfdfc.tar.xz"
    version = "3.1.7"

@@ -42,7 +42,7 @@ api = "0.2"
    sha256 = "8dcf99567d40190c69c875847f7ed9c7158ad78643a17089775ac0097965f09e"
    source = "https://download.visualstudio.microsoft.com/download/pr/f7c8f82a-8c47-497d-875b-2ac210599ec5/e8aea0c195efed8a9aff2ba687db8c26/aspnetcore-runtime-3.1.8-linux-x64.tar.gz"
    source_sha256 = "823f8ea555fd56ab40d56d423748036204c4540c08baa61de4462978a0c35583"
-    stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.cflinuxfs3"]
+    stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.cflinuxfs3", "io.paketo.stacks.tiny"]
    uri = "https://buildpacks.cloudfoundry.org/dependencies/dotnet-aspnetcore/dotnet-aspnetcore_3.1.8_linux_x64_any-stack_8dcf9956.tar.xz"
    version = "3.1.8"

@@ -51,3 +51,6 @@ api = "0.2"

[[stacks]]
  id = "org.cloudfoundry.stacks.cflinuxfs3"
+
+[[stacks]]
+  id = "io.paketo.stacks.tiny"
```

### dotnet-core-build
1. Add `io.paketo.stacks.tiny` to the `stacks` list in the `buildpack.toml`
   ```diff
   diff --git a/buildpack.toml b/buildpack.toml
   index 1189416..4733c64 100644
   --- a/buildpack.toml
   +++ b/buildpack.toml
   @@ -15,3 +15,6 @@ id = "org.cloudfoundry.stacks.cflinuxfs3"

    [[stacks]]
    id = "io.buildpacks.stacks.bionic"
   +
   +[[stacks]]
   +id = "io.paketo.stacks.tiny"
   ```

1. Include check for `io.paketo.stacks.tiny` in `detect.go` to include `icu` in the buildplan
   ```diff
   diff --git a/cmd/detect/main.go b/cmd/detect/main.go
   index fb7a748..ee68a96 100644
   --- a/cmd/detect/main.go
   +++ b/cmd/detect/main.go
   @@ -104,7 +104,7 @@ func runDetect(context detect.Detect) (int, error) {
       })
     }

   -	if context.Stack == "io.buildpacks.stacks.bionic" {
   +	if context.Stack == "io.buildpacks.stacks.bionic" || context.Stack == "io.paketo.stacks.tiny" {
       plan.Requires = append(plan.Requires, buildplan.Required{
         Name:     "icu",
         Metadata: buildplan.Metadata{"build": true},
   ```

### dotnet-core-conf
1. Add `io.paketo.stacks.tiny` to the `stacks` list in the `buildpack.toml`
   ```diff
   diff --git a/buildpack.toml b/buildpack.toml
   index 92b8957..29c6184 100644
   --- a/buildpack.toml
   +++ b/buildpack.toml
   @@ -1,4 +1,4 @@
   -api = "0.2"
   +api = "0.4"

    [buildpack]
    id = "paketo-buildpacks/dotnet-core-conf"
   @@ -15,3 +15,6 @@ id = "org.cloudfoundry.stacks.cflinuxfs3"

    [[stacks]]
    id = "io.buildpacks.stacks.bionic"
   +
   +[[stacks]]
   +id = "io.paketo.stacks.tiny"
   ```

1. Include check for `io.paketo.stacks.tiny` in `detect.go` to include `icu` in the buildplan
   ```diff
   diff --git a/cmd/detect/main.go b/cmd/detect/main.go
   index 3f311e5..e3d180d 100644
   --- a/cmd/detect/main.go
   +++ b/cmd/detect/main.go
   @@ -61,7 +61,7 @@ func runDetect(context detect.Detect) (int, error) {
    		return context.Fail(), nil
    	}

   -	if context.Stack == "io.buildpacks.stacks.bionic" {
   +	if context.Stack == "io.buildpacks.stacks.bionic" || context.Stack == "io.paketo.stacks.tiny" {
    		plan.Requires = append(plan.Requires, buildplan.Required{
    			Name:     "icu",
    			Metadata: buildplan.Metadata{"launch": true},
   ```

1. Rewrite how the start command is generated as `io.paketo.stacks.tiny` does not have a shell.
   ```diff
   diff --git a/conf/conf.go b/conf/conf.go
   index d498d51..7621176 100644
   --- a/conf/conf.go
   +++ b/conf/conf.go
   @@ -38,20 +38,21 @@ func (c Contributor) Contribute() error {
    		return err
    	}

   -	startCmdPrefix := fmt.Sprintf("dotnet %s.dll", runtimeConfig.BinaryName)
   +	command := "dotnet"
   +	var args []string
    	if hasExecutable {
   -		startCmdPrefix = fmt.Sprintf("./%s", runtimeConfig.BinaryName)
   +		command = fmt.Sprintf("./%s", runtimeConfig.BinaryName)
   +	} else {
   +		args = append(args, fmt.Sprintf("%s.dll", runtimeConfig.BinaryName))
    	}

   -	args := fmt.Sprintf("%s --urls http://0.0.0.0:${PORT:-8080}", startCmdPrefix)
   -	startCmd := fmt.Sprintf("cd %s && %s", c.context.Application.Root, args)
   -
    	return c.context.Layers.WriteApplicationMetadata(layers.Metadata{
    		Processes: []layers.Process{
    			{
    				Type:    "web",
   -				Command: startCmd,
   -				Direct:  false,
   +				Command: command,
   +				Args:    args,
   +				Direct:  true,
    			},
    		},
    	})
   ```

### dotnet-core-runtime
1. Add `io.paketo.stacks.tiny` to the `stacks` list in the `buildpack.toml`
1. Include `io.paketo.stacks.tiny` on the list of `stacks` for all `metadata.dependencies` in the `buildpack.toml`
```diff
diff --git a/buildpack.toml b/buildpack.toml
index 3cf27ad..6d9a734 100644
--- a/buildpack.toml
+++ b/buildpack.toml
@@ -16,7 +16,7 @@ api = "0.2"
    sha256 = "e88b8aadaf0e4feebd508aca33e35d149d2ef9443cf151b90d2882e9afa230dc"
    source = "https://download.visualstudio.microsoft.com/download/pr/76cf51d4-8407-46a9-9ba0-c44b8c62b553/8af610974c8636cd4e7b7ec0f17ac32a/dotnet-runtime-2.1.21-linux-x64.tar.gz"
    source_sha256 = "58a4a3f4fdb529db928f586cd8267654db640d41fa4a88b270f7fdec25a25889"
-    stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.cflinuxfs3"]
+    stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.cflinuxfs3", "io.paketo.stacks.tiny"]
    uri = "https://buildpacks.cloudfoundry.org/dependencies/dotnet-runtime/dotnet-runtime_2.1.21_linux_x64_any-stack_e88b8aad.tar.xz"
    version = "2.1.21"

@@ -26,7 +26,7 @@ api = "0.2"
    sha256 = "81711e4edea078a4115a67e908591f813ee22b7934f735d7bd6e9e15e906bdfb"
    source = "https://download.visualstudio.microsoft.com/download/pr/926c221c-a9bd-4022-a0bd-52f93d273883/a8582353d501c69bd991c52a52d79bae/dotnet-runtime-2.1.22-linux-x64.tar.gz"
    source_sha256 = "d4faaaed24b9bf5afaa6a777343dccbd6a05f267541b857d02ca16146dc54a2d"
-    stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.cflinuxfs3"]
+    stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.cflinuxfs3", "io.paketo.stacks.tiny"]
    uri = "https://buildpacks.cloudfoundry.org/dependencies/dotnet-runtime/dotnet-runtime_2.1.22_linux_x64_any-stack_81711e4e.tar.xz"
    version = "2.1.22"

@@ -36,7 +36,7 @@ api = "0.2"
    sha256 = "52ccd274b71c6dd8eefbb0b8a16e45cf9997af96e71ea6b7103ddd9e70f3261c"
    source = "https://download.visualstudio.microsoft.com/download/pr/e42ed5c3-d7a3-404d-a242-cfd10ef626ff/b723e456ffaf60b6df6c6d5b0a792aba/dotnet-runtime-3.1.7-linux-x64.tar.gz"
    source_sha256 = "51c719a8c085baaeca9eef0cdb5a0a0cb8a15ef73f4cf0688d751a12a8b1df41"
-    stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.cflinuxfs3"]
+    stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.cflinuxfs3", "io.paketo.stacks.tiny"]
    uri = "https://buildpacks.cloudfoundry.org/dependencies/dotnet-runtime/dotnet-runtime_3.1.7_linux_x64_any-stack_52ccd274.tar.xz"
    version = "3.1.7"

@@ -46,7 +46,7 @@ api = "0.2"
    sha256 = "a1e739c553f61337bcd642ba3077047628f104023116c4eac1587fd00426ea3f"
    source = "https://download.visualstudio.microsoft.com/download/pr/e4e47a0a-132e-416a-b8eb-f3373ad189d9/43af4412e27696c3c16e50f496f6c7af/dotnet-runtime-3.1.8-linux-x64.tar.gz"
    source_sha256 = "c50800e02cea23609ec6a009b1fbfe6b1f7ec4634c54bee089f918fca8fe2323"
-    stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.cflinuxfs3"]
+    stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.cflinuxfs3", "io.paketo.stacks.tiny"]
    uri = "https://buildpacks.cloudfoundry.org/dependencies/dotnet-runtime/dotnet-runtime_3.1.8_linux_x64_any-stack_a1e739c5.tar.xz"
    version = "3.1.8"

@@ -67,3 +67,6 @@ api = "0.2"

[[stacks]]
  id = "org.cloudfoundry.stacks.cflinuxfs3"
+
+[[stacks]]
+  id = "io.paketo.stacks.tiny"
```

### dotnet-core-sdk
1. Add `io.paketo.stacks.tiny` to the `stacks` list in the `buildpack.toml`
1. Include `io.paketo.stacks.tiny` on the list of `stacks` for all `metadata.dependencies` in the `buildpack.toml`
```diff
diff --git a/buildpack.toml b/buildpack.toml
index e75aa19..1b29d43 100644
--- a/buildpack.toml
+++ b/buildpack.toml
@@ -15,7 +15,7 @@ api = "0.2"
    sha256 = "849799474b03d2f722170bf2fff6dc8bb08ca4ebc10c86774559f9d1a4deb1bc"
    source = "https://download.visualstudio.microsoft.com/download/pr/a44fb0b1-2c91-41d6-8970-321872341326/7e150d5bc0d3d96ae8c7cbd9e6b890fe/dotnet-sdk-2.1.809-linux-x64.tar.gz"
    source_sha256 = "0c79f6133aa3394b683978774e0975122cd9d58b90c3b0b65bba48d44f5bafc0"
-    stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.cflinuxfs3"]
+    stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.cflinuxfs3", "io.paketo.stacks.tiny"]
    uri = "https://buildpacks.cloudfoundry.org/dependencies/dotnet-sdk/dotnet-sdk_2.1.809_linux_x64_any-stack_84979947.tar.xz"
    version = "2.1.809"

@@ -25,7 +25,7 @@ api = "0.2"
    sha256 = "fde667012629b99f2e093caf736e2a8b50dc3206e1062a2622abb526d1ea08b6"
    source = "https://download.visualstudio.microsoft.com/download/pr/eb1b19f5-3c42-4f7b-b36a-67fae2040506/40cc70f95b6485b0b87bcbc655b7c855/dotnet-sdk-2.1.810-linux-x64.tar.gz"
    source_sha256 = "3856c888ed777818f6e4fb38434adbd139abd0cd4512c0579847503238987b65"
-    stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.cflinuxfs3"]
+    stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.cflinuxfs3", "io.paketo.stacks.tiny"]
    uri = "https://buildpacks.cloudfoundry.org/dependencies/dotnet-sdk/dotnet-sdk_2.1.810_linux_x64_any-stack_fde66701.tar.xz"
    version = "2.1.810"

@@ -34,7 +34,7 @@ api = "0.2"
    sha256 = "94ec0b48b052227519386233aeb84521440753951bbba6713702918bf6d71012"
    source = "https://download.visualstudio.microsoft.com/download/pr/4f9b8a64-5e09-456c-a087-527cfc8b4cd2/15e14ec06eab947432de139f172f7a98/dotnet-sdk-3.1.401-linux-x64.tar.gz"
    source_sha256 = "292d8f5694df7560c39a16c12d5b5efa4038c0973d1adb768f90f39982da1c43"
-    stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.cflinuxfs3"]
+    stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.cflinuxfs3", "io.paketo.stacks.tiny"]
    uri = "https://buildpacks.cloudfoundry.org/dependencies/dotnet-sdk/dotnet-sdk_3.1.401_linux_x64_any-stack_94ec0b48.tar.xz"
    version = "3.1.401"

@@ -43,7 +43,7 @@ api = "0.2"
    sha256 = "e0aedde79c13a4a58e0fb85dc7d12fe005675a4214bec009680d412981ece15a"
    source = "https://download.visualstudio.microsoft.com/download/pr/f01e3d97-c1c3-4635-bc77-0c893be36820/6ec6acabc22468c6cc68b61625b14a7d/dotnet-sdk-3.1.402-linux-x64.tar.gz"
    source_sha256 = "2b6b172f9483e499141e37a6b932a547d9476bf03f3e71a0fefb76c52e01a9ee"
-    stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.cflinuxfs3"]
+    stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.cflinuxfs3", "io.paketo.stacks.tiny"]
    uri = "https://buildpacks.cloudfoundry.org/dependencies/dotnet-sdk/dotnet-sdk_3.1.402_linux_x64_any-stack_e0aedde7.tar.xz"
    version = "3.1.402"

@@ -68,3 +68,6 @@ api = "0.2"

[[stacks]]
  id = "org.cloudfoundry.stacks.cflinuxfs3"
+
+[[stacks]]
+  id = "io.paketo.stacks.tiny"
```

### icu
1. Add `io.paketo.stacks.tiny` to the `stacks` list in the `buildpack.toml`
1. Include `io.paketo.stacks.tiny` on the list of `stacks` for all `metadata.dependencies` in the `buildpack.toml`
```diff
diff --git a/buildpack.toml b/buildpack.toml
index 38b9b3f..c21a774 100644
--- a/buildpack.toml
+++ b/buildpack.toml
@@ -15,7 +15,7 @@ api = "0.2"
    sha256 = "b31f08f61f93fd361bceeed815119cf2108d24228479abd2639ceb09a9d71b88"
    source = "https://github.com/unicode-org/icu/releases/download/release-66-1/icu4c-66_1-Ubuntu18.04-x64.tgz"
    source_sha256 = "20c995f4d1285b31cf6aace2d3e7fe12bd974e3cb6776f7a174e5976d6b88fec"
-    stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.cflinuxfs3"]
+    stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.cflinuxfs3", "io.paketo.stacks.tiny"]
    uri = "https://buildpacks.cloudfoundry.org/dependencies/icu/icu-66.1.0-any-stack-b31f08f6.tgz"
    version = "66.1.0"

@@ -24,7 +24,7 @@ api = "0.2"
    sha256 = "00267b036b85449b730ccca3b18d528e13a207b88c8b43f6a6edca1dc21abe31"
    source = "https://github.com/unicode-org/icu/releases/download/release-67-1/icu4c-67_1-Ubuntu18.04-x64.tgz"
    source_sha256 = "303e71ecb746b767a0e899ef6e3733c0902d8f211d5fc660f8b0524d7e791ccb"
-    stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.cflinuxfs3"]
+    stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.cflinuxfs3", "io.paketo.stacks.tiny"]
    uri = "https://buildpacks.cloudfoundry.org/dependencies/icu/icu_67.1.0_linux_noarch_any-stack_00267b03.tgz"
    version = "67.1.0"

@@ -33,3 +33,6 @@ api = "0.2"

[[stacks]]
  id = "org.cloudfoundry.stacks.cflinuxfs3"
+
+[[stacks]]
+  id = "io.paketo.stacks.tiny"
```

### node-engine
1. Add `io.paketo.stacks.tiny` to the `stacks` list in the `buildpack.toml`
1. Include `io.paketo.stacks.tiny` on the list of `stacks` for all `metadata.dependencies` in the `buildpack.toml`
```diff
diff --git a/buildpack.toml b/buildpack.toml
index 0c8d949..d51a608 100644
--- a/buildpack.toml
+++ b/buildpack.toml
@@ -18,7 +18,7 @@ api = "0.2"
    sha256 = "43616969dd39d52c9d3c0a4ed5e66600133356877ab8344a916638e7f5794490"
    source = "https://nodejs.org/dist/v10.22.0/node-v10.22.0.tar.gz"
    source_sha256 = "8a77f883a9cba5451cef547f737e590a32c9840a4ab421a048f2fadda799ba41"
-    stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.cflinuxfs3"]
+    stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.cflinuxfs3", "io.paketo.stacks.tiny"]
    uri = "https://buildpacks.cloudfoundry.org/dependencies/node/node_10.22.0_linux_x64_cflinuxfs3_43616969.tgz"
    version = "10.22.0"

@@ -29,7 +29,7 @@ api = "0.2"
    sha256 = "d84ec1b77780923f2d0d30f9155dfcba411c67ae53548684a140b2ec982fdeba"
    source = "https://nodejs.org/dist/v10.22.1/node-v10.22.1.tar.gz"
    source_sha256 = "d0b49dd96ac70e99240458863efe09ae5bb1138c0ff582295f882c1482708172"
-    stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.cflinuxfs3"]
+    stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.cflinuxfs3", "io.paketo.stacks.tiny"]
    uri = "https://buildpacks.cloudfoundry.org/dependencies/node/node_10.22.1_linux_x64_cflinuxfs3_d84ec1b7.tgz"
    version = "10.22.1"

@@ -40,7 +40,7 @@ api = "0.2"
    sha256 = "760e26561e981223ff92b666bbab6bf66b2ae652b42bd2bb5dc6c8163b5e28fe"
    source = "https://nodejs.org/dist/v12.18.3/node-v12.18.3.tar.gz"
    source_sha256 = "6ea85f80e01b007cc9b566b8836513bc5102667d833bad4c1092be60fa60c2d4"
-    stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.cflinuxfs3"]
+    stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.cflinuxfs3", "io.paketo.stacks.tiny"]
    uri = "https://buildpacks.cloudfoundry.org/dependencies/node/node_12.18.3_linux_x64_cflinuxfs3_760e2656.tgz"
    version = "12.18.3"

@@ -51,7 +51,7 @@ api = "0.2"
    sha256 = "2c0046f5d0bdccf7738ba6b5e3f1084a866ff4a92c5a3f3e820b3fd8e2101fbc"
    source = "https://nodejs.org/dist/v12.18.4/node-v12.18.4.tar.gz"
    source_sha256 = "a802d87e579e46fc52771ed6f2667048320caca867be3276f4c4f1bbb41389c3"
-    stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.cflinuxfs3"]
+    stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.cflinuxfs3", "io.paketo.stacks.tiny"]
    uri = "https://buildpacks.cloudfoundry.org/dependencies/node/node_12.18.4_linux_x64_cflinuxfs3_2c0046f5.tgz"
    version = "12.18.4"

@@ -62,7 +62,7 @@ api = "0.2"
    sha256 = "9b95bd4a9d3b933e10350586dabfa3ce61ebfa1e39adaec7d86c586b3c7feead"
    source = "https://nodejs.org/dist/v14.10.1/node-v14.10.1.tar.gz"
    source_sha256 = "5047c4962012f88258d8c1c6c133d870fd818ed5ea0f194ab3aa206510d144ae"
-    stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.cflinuxfs3"]
+    stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.cflinuxfs3", "io.paketo.stacks.tiny"]
    uri = "https://buildpacks.cloudfoundry.org/dependencies/node/node_14.10.1_linux_x64_cflinuxfs3_9b95bd4a.tgz"
    version = "14.10.1"

@@ -73,7 +73,7 @@ api = "0.2"
    sha256 = "c81330009f27f95a3cf41003c290692efcb8d2b89a1028d7e9d01fb0de79c181"
    source = "https://nodejs.org/dist/v14.11.0/node-v14.11.0.tar.gz"
    source_sha256 = "c07669ddbd708d0dfc4ccb63a7ced7ad1fd7d1b59ced50cf05f22f0b96e45463"
-    stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.cflinuxfs3"]
+    stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.cflinuxfs3", "io.paketo.stacks.tiny"]
    uri = "https://buildpacks.cloudfoundry.org/dependencies/node/node_14.11.0_linux_x64_cflinuxfs3_c8133000.tgz"
    version = "14.11.0"

@@ -100,3 +100,6 @@ api = "0.2"

[[stacks]]
  id = "org.cloudfoundry.stacks.cflinuxfs3"
+
+[[stacks]]
+  id = "io.paketo.stacks.tiny"
```

## Stack Changes

In order to get a .Net Core app to run on top of the `io.paketo.stacks.tiny` Run Image, we need to add a couple of packages:
1. libstdc++6
1. libgcc1
```diff
diff --git a/tiny/dockerfile/run/packagelist b/tiny/dockerfile/run/packagelist
index 55055e4..dc9e1bd 100644
--- a/tiny/dockerfile/run/packagelist
+++ b/tiny/dockerfile/run/packagelist
@@ -1,7 +1,9 @@
base-files
ca-certificates
libc6
+libgcc1
libssl1.1
+libstdc++6
netbase
openssl
tzdata
```
