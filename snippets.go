//Log into the registry

var authConfig = registry.AuthConfig{
Username:      "gsaenger",
Password:      os.Getenv("DOCKER_PASS"),
ServerAddress: "https://index.docker.io/v1/",
}

authConfigBytes, _ := json.Marshal(authConfig)
authConfigEncoded := base64.URLEncoding.EncodeToString(authConfigBytes)

// Push image

tag := "gsaenger/hello-go"
pushOpts := types.ImagePushOptions{RegistryAuth: authConfigEncoded}
pushLogs, err := dockerClient.ImagePush(context.Background(), tag, pushOpts)
if err != nil {
fmt.Println(err.Error())
return
}

defer pushLogs.Close()

err = printOutput(pushLogs)
if err != nil {
fmt.Println(err.Error())
return
}



// Start a session for BuildKit
cfg, err := getDefaultDockerConfig()
ctx := context.Background()

sess, err := session.NewSession(ctx, "pulumi-docker", identity.NewID())
if err != nil {
fmt.Println(err.Error())
return
}

dockerAuthProvider := authprovider.NewDockerAuthProvider(cfg, nil)
sess.Allow(dockerAuthProvider)

dialSession := func(ctx context.Context, proto string, meta map[string][]string) (net.Conn, error) {
return dockerClient.DialHijack(ctx, "/session", proto, meta)
}
go func() {
err := sess.Run(ctx, dialSession)
if err != nil {
fmt.Println(err.Error())
return
}
}()
defer sess.Close()
opts.SessionID = sess.ID()


dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
if err != nil {
fmt.Println(err.Error())
return
}