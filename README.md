http-ldap-authrequest
=====================

Description
-----------
http-ldap-authrequest is an implementation of LDAP authentification service similar to the ldap-auth daemon described in [reference nginx-ldap-auth](https://github.com/nginxinc/nginx-ldap-auth). Although the implementation is intended to be used as the responder to nginx [auth_request](https://nginx.org/en/docs/http/ngx_http_auth_request_module.html) it should work with other products, for example as [traefik ForwardAuth](https://doc.traefik.io/traefik/middlewares/http/forwardauth/).

http-ldap-authrequest expects `Basic` scheme in `Authorization` header.

Configuration
-------------

### Simple

example:
```
http-ldap-authrequest -loglevel=debug -ldapuri=ldaps://ldap.example.com  \
    -auth-bind-dn=cn=iam,ou=ims,dc=example,dc=local
    -auth-bind-pw=./bindpass
    -user-base-dn=ou=users,ou=iam,ou=dc=example,dc=local
    -group-base-dn=ou=groups,ou=iam,dc=example,dc=local
    -group-attr=cn
    -group-filter=(uniqueMember=%s)
```

### Kubernetes ingress-nginx

You may add `http-ldap-authrequest` as a side car container to you Kubernetes ingress-nginx to utilize LDAP auth. Using `ingress-nginx` Helm chart  `http-ldap-authrequest` can be added as a side car to your nginx-ingress and configured as follows (just the part related to the side car):

```
controller:
  extraContainers:
  - image: http-ldap-authrequest:latest
    name: http-ldap-authrequest
    args:
    - "-loglevel=debug"
    - "-ldapuri=ldaps://ldap.example.com"
    - "-auth-bind-pw=/var/run/secrets/http-ldap-authrequest/bindpass"
    - "-auth-bind-dn=cn=admin,dc=example,dc=local"
    - "-user-base-dn=dc=example,dc=local"
    - "-group-base-dn=ou=groups,dc=example,dc=local"
    - "-group-attr=cn"
    - "-group-filter=(uniqueMember=%s)"
    ports:
    - name: http
      containerPort: 8242
    volumeMounts:
    - name: http-ldap-authrequest-secret-config
      mountPath: /var/run/secrets/http-ldap-authrequest
    resources:
      limits:
        cpu: 50m
        memory: 20Mi
      requests:
        cpu: 10m
        memory: 5Mi

  extraVolumes:
  - name: http-ldap-authrequest-secret-config
    secret:
      secretName: http-ldap-authrequest
```

where secret http-ldap-authrequest contains bind password for the authentificator.

You ingress manifest may look as following:
```
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: myservice
  annotations:
    kubernetes.io/ingress.class: "nginx"
    nginx.ingress.kubernetes.io/auth-url: http://localhost:8242
    nginx.ingress.kubernetes.io/auth-cache-key: "$remote_user$http_authorization"
    nginx.ingress.kubernetes.io/auth-cache-duration: "200 202 401 1m"
    nginx.ingress.kubernetes.io/auth-snippet: |
      proxy_set_header X-Http-Ldap-Authrequest-RequiredGroup mygroup;
spec:
  rules:
  - host: myservice.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: myservice
            port:
              number: 80
```

where
- `nginx.ingress.kubernetes.io/auth-url` points to http-ldap-authrequest side car
- optional `nginx.ingress.kubernetes.io/auth-cache-key` enables cach and `nginx.ingress.kubernetes.io/auth-cache-duration` allows caching of auth in nginx, see [details](https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/annotations/#external-authentication).
- `nginx.ingress.kubernetes.io/auth-snippet` passes additionally header X-Http-Ldap-Authrequest-RequiredGroup: mygroup, which makes http-ldap-authrequest additionally to check if the authenticated user is also a member of the specified group `mygroup`.

If your backend does not perform additional authorization, you may want to remove `Authorization` header from the requests forwarded to the backend. This is especially important, if your backend operates over http (i.e. traffic is not encrypted):
```
  annotations:
    nginx.ingress.kubernetes.io/configuration-snippet: |
      proxy_set_header authorization "";
```

Build
-----
```
make
```

```
docker build -t http-ldap-authrequest:$(git describe) .
```
