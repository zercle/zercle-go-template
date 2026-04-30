# Kustomize Migration Plan: Docker Compose to Kubernetes

## Status
**Planning Phase** - Ready for implementation

## Goal
Migrate the Docker Compose-based deployment to a complete Kustomize-based Kubernetes deployment that correctly handles the multi-binary application architecture (server, client, migrate) with proper environment variable naming matching the viper config loading.

---

## REASONS Canvas

### R — Requirements

**Problem:** The existing Kustomize setup is incomplete and incorrect:
1. Only has ONE deployment instead of separate server/client deployments
2. Missing PostgreSQL and Valkey StatefulSets with persistent storage
3. ConfigMap uses WRONG env var names (no `APP_` prefix, wrong key names)
4. No separation between gRPC server and HTTP client concerns

**Definition of Done:**
- Kubernetes manifests deploy the full stack: PostgreSQL, Valkey, Server (gRPC+HTTP), Client (HTTP+SSE), Migration Job
- All env vars correctly prefixed with `APP_` and matching viper config structure
- Dev/Prod overlays correctly patch the appropriate resources
- Health checks, HPA, PDB configured for both server and client
- Ingress routes to the client service (external-facing API)

**Acceptance Criteria:**

```gherkin
Given a Kubernetes cluster
When I run `kubectl apply -k deployments/kustomize/overlays/development`
Then all resources should be created successfully
And pods should be in Running state
And the server should listen on gRPC port 50051
And the client should respond on HTTP port 8080
And the /health endpoint should return 200 OK
And the /ready endpoint should return 200 OK with database and valkey status

Given the deployed application
When I check environment variables in server pod
Then APP_SERVER_GRPC_HOST should be set to 0.0.0.0
And APP_DATABASE_HOST should point to postgres service
And APP_VALKEY_HOST should point to valkey service

Given the deployed application
When I check environment variables in client pod
Then APP_SERVER_HTTP_HOST should be set to 0.0.0.0
And APP_DATABASE_HOST should point to postgres service
And APP_VALKEY_HOST should point to valkey service
```

### E — Entities

**Domain Entities:**

| Entity | Type | Description |
|--------|------|-------------|
| `postgres` | StatefulSet | PostgreSQL 18-alpine with persistent storage |
| `valkey` | StatefulSet | Valkey 9 with persistent storage |
| `server` | Deployment | gRPC server (50051) + HTTP Echo (8080), needs postgres |
| `client` | Deployment | HTTP Echo (8080) with SSE, needs postgres + valkey |
| `migrate` | Job | Database migration runner, runs once per deploy |
| `config` | ConfigMap | Non-sensitive configuration with APP_* keys |
| `secrets` | Secret | Sensitive credentials with APP_* keys |
| `ingress` | Ingress | Routes external traffic to client service |

**Resource Relationships:**

```
┌─────────────────────────────────────────────────────────────┐
│                         Ingress                              │
│              Routes: / → client-service:8080                │
└───────────────────────┬─────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────────┐
│                     client-service                          │
│              Selector: app=client                           │
│              Ports: 8080 (http)                             │
└───────────────────────┬─────────────────────────────────────┘
                        │
        ┌───────────────┼───────────────┐
        ▼               ▼               ▼
┌──────────────┐ ┌──────────────┐ ┌──────────────┐
│   Client     │ │   Server     │ │   Migrate    │
│   Pod(s)     │ │   Pod(s)     │ │   Job        │
│              │ │              │ │              │
│ Env:         │ │ Env:         │ │ Env:         │
│ APP_DATABASE │ │ APP_DATABASE │ │ APP_DATABASE │
│ APP_VALKEY   │ │ APP_VALKEY   │ │              │
└──────┬───────┘ └──────┬───────┘ └──────┬───────┘
       │                │                │
       └────────────────┼────────────────┘
                        │
                        ▼
           ┌────────────────────────┐
           │    postgres-service    │
           │    Port: 5432          │
           └───────────┬────────────┘
                       │
           ┌───────────┴───────────┐
           ▼                       ▼
   ┌───────────────┐       ┌───────────────┐
   │ postgres-0    │       │   postgres    │
   │ (StatefulSet) │       │   (PVC)       │
   └───────────────┘       └───────────────┘

           ┌────────────────────────┐
           │    valkey-service      │
           │    Port: 6379          │
           └───────────┬────────────┘
                       │
           ┌───────────┴───────────┐
           ▼                       ▼
   ┌───────────────┐       ┌───────────────┐
   │ valkey-0      │       │   valkey      │
   │ (StatefulSet) │       │   (PVC)       │
   └───────────────┘       └───────────────┘
```

### A — Approach

**Strategy:**
1. **Separate Concerns**: Split the monolithic deployment into separate server and client deployments with distinct responsibilities
2. **Infrastructure-First**: Create StatefulSets for PostgreSQL and Valkey before application deployments
3. **Config Correction**: Fix all environment variable names to match viper's `APP_` prefix requirement and config structure
4. **Service Mesh**: Create appropriate Services for inter-pod communication (postgres, valkey, server-gRPC, client-http)
5. **Migration Order**: Migration Job runs before application pods (via init containers or helm-style hooks)

**Design Decisions:**

| Decision | Rationale |
|----------|-----------|
| StatefulSet for PostgreSQL | Requires persistent storage, ordered deployment, stable network identity |
| StatefulSet for Valkey | Requires persistent storage, single instance is sufficient |
| Separate Deployments | Server and client have different scaling needs, resource profiles, and concerns |
| ConfigMap with literals | Simplifies Kustomize merging; literals merge better than files |
| Init container for migrations | Ensures migrations run before app starts without external dependencies |
| HPA per deployment | Server and client may have different scaling triggers |

**Alternatives Considered:**
- **Helm vs Kustomize**: Keeping Kustomize for simplicity and native Kubernetes integration
- **Single Deployment with two containers**: Rejected - different scaling needs, separate binaries
- **External PostgreSQL/Valkey**: Rejected - self-contained for dev, can be swapped in prod via overlays

### S — Structure

**File Structure:**

```
deployments/kustomize/
├── base/
│   ├── kustomization.yaml           # UPDATED: new resources list, correct images
│   ├── namespace.yaml               # KEEP: namespace definition
│   ├── serviceaccount.yaml          # KEEP: SA for pods
│   │
│   ├── postgres-statefulset.yaml    # NEW: PostgreSQL StatefulSet + Service + PVC
│   ├── valkey-statefulset.yaml      # NEW: Valkey StatefulSet + Service + PVC
│   │
│   ├── server-deployment.yaml       # NEW: Server deployment (gRPC 50051 + HTTP 8080)
│   ├── server-service.yaml          # NEW: Server service (gRPC + HTTP ports)
│   ├── server-hpa.yaml              # NEW: Server HPA
│   ├── server-pdb.yaml              # NEW: Server PDB
│   │
│   ├── client-deployment.yaml       # NEW: Client deployment (HTTP 8080)
│   ├── client-service.yaml          # NEW: Client service (HTTP port)
│   ├── client-hpa.yaml              # NEW: Client HPA
│   ├── client-pdb.yaml              # NEW: Client PDB
│   │
│   ├── configmap.yaml               # NEW: ConfigMap with correct APP_* env vars
│   ├── secret.yaml                  # NEW: Secret with correct APP_* env vars
│   ├── ingress.yaml                 # UPDATED: point to client-service
│   └── migration-job.yaml           # UPDATED: use correct env vars
│
├── overlays/
│   ├── development/
│   │   ├── kustomization.yaml       # UPDATED: new patches, configmap merge
│   │   └── patches/
│   │       ├── server-deployment-patch.yaml   # NEW
│   │       ├── client-deployment-patch.yaml   # NEW
│   │       └── deployment-patch.yaml          # DELETE
│   │
│   └── production/
│       ├── kustomization.yaml       # UPDATED: new patches, configmap merge
│       └── patches/
│           ├── server-deployment-patch.yaml   # NEW
│           ├── client-deployment-patch.yaml   # NEW
│           └── deployment-patch.yaml          # DELETE
```

**Dependency Graph:**

```
                                    ┌─────────────────┐
                                    │   namespace     │
                                    │   serviceaccount│
                                    └────────┬────────┘
                                             │
            ┌────────────────────────────────┼────────────────────────────────┐
            │                                │                                │
            ▼                                ▼                                ▼
   ┌─────────────────┐            ┌─────────────────┐            ┌─────────────────┐
   │ postgres-ss     │            │ valkey-ss       │            │ config/secret   │
   │ (infrastructure)│            │ (infrastructure)│            │ (configuration) │
   └────────┬────────┘            └────────┬────────┘            └────────┬────────┘
            │                              │                              │
            └──────────────┬───────────────┘                              │
                           │                                              │
                           ▼                                              ▼
                  ┌─────────────────┐                           ┌─────────────────┐
                  │ migration-job   │◄──────────────────────────┤ requires config │
                  │ (pre-req)       │                           │ and secrets     │
                  └────────┬────────┘                           └─────────────────┘
                           │
            ┌──────────────┼──────────────┐
            │              │              │
            ▼              ▼              ▼
   ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
   │ server-deploy   │ │ client-deploy   │ │ ingress         │
   │ (depends on     │ │ (depends on     │ │ (routes to      │
   │  migration)     │ │  migration)     │ │  client)        │
   └────────┬────────┘ └────────┬────────┘ └─────────────────┘
            │                   │
            ▼                   ▼
   ┌─────────────────┐ ┌─────────────────┐
   │ server-hpa      │ │ client-hpa      │
   │ server-pdb      │ │ client-pdb      │
   └─────────────────┘ └─────────────────┘
```

### O — Operations

#### Task 1: Create Infrastructure StatefulSets [P]

**T1.1: Create PostgreSQL StatefulSet**
- **File:** `deployments/kustomize/base/postgres-statefulset.yaml`
- **Content:**
  - Headless Service: `postgres-service`, port 5432
  - StatefulSet: `postgres`, 1 replica
  - Container: `postgres:18-alpine`
  - Env: `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`, `TZ`
  - VolumeClaimTemplate: 10Gi storage
  - Resources: requests 100m/256Mi, limits 500m/512Mi
  - LivenessProbe: `pg_isready -U postgres`
  - ReadinessProbe: `pg_isready -U postgres`

**T1.2: Create Valkey StatefulSet**
- **File:** `deployments/kustomize/base/valkey-statefulset.yaml`
- **Content:**
  - Headless Service: `valkey-service`, port 6379
  - StatefulSet: `valkey`, 1 replica
  - Container: `valkey/valkey:9`
  - Env: `TZ`
  - VolumeClaimTemplate: 5Gi storage
  - Resources: requests 50m/128Mi, limits 200m/256Mi
  - LivenessProbe: `valkey-cli ping`
  - ReadinessProbe: `valkey-cli ping`

#### Task 2: Create Configuration Resources [P]

**T2.1: Create ConfigMap with Correct Env Vars**
- **File:** `deployments/kustomize/base/configmap.yaml`
- **Keys (all prefixed implicitly by viper, but keys here match config structure):**
  ```yaml
  APP_SERVER_GRPC_HOST: "0.0.0.0"
  APP_SERVER_GRPC_PORT: "50051"
  APP_SERVER_HTTP_HOST: "0.0.0.0"
  APP_SERVER_HTTP_PORT: "8080"
  APP_DATABASE_HOST: "postgres-service"
  APP_DATABASE_PORT: "5432"
  APP_DATABASE_NAME: "zercle_chat"
  APP_DATABASE_SSL_MODE: "disable"
  APP_DATABASE_MAX_CONNECTIONS: "25"
  APP_DATABASE_MAX_IDLE_CONNS: "5"
  APP_VALKEY_HOST: "valkey-service"
  APP_VALKEY_PORT: "6379"
  APP_VALKEY_DB: "0"
  APP_LOGGING_LEVEL: "info"
  APP_LOGGING_FORMAT: "json"
  ```

**T2.2: Create Secret with Sensitive Values**
- **File:** `deployments/kustomize/base/secret.yaml`
- **Keys:**
  ```yaml
  APP_DATABASE_USER: "postgres"
  APP_DATABASE_PASSWORD: "postgres"
  APP_VALKEY_PASSWORD: ""  # empty for dev
  APP_AUTH_JWT_SECRET: "your-secret-key-change-in-production"
  ```

#### Task 3: Create Server Deployment Resources

**T3.1: Create Server Deployment**
- **File:** `deployments/kustomize/base/server-deployment.yaml`
- **Content:**
  - Deployment: `server`, 2 replicas
  - Labels: `app: server`, `component: backend`
  - Container: `zercle-go-template-server:latest`
  - Ports: 50051 (grpc), 8080 (http)
  - EnvFrom: configmap, secret
  - Resources: requests 100m/128Mi, limits 500m/512Mi
  - LivenessProbe: gRPC health check or HTTP /healthz on 8080
  - ReadinessProbe: gRPC health check or HTTP /readyz on 8080
  - SecurityContext: runAsNonRoot, runAsUser 65534

**T3.2: Create Server Service**
- **File:** `deployments/kustomize/base/server-service.yaml`
- **Content:**
  - Service: `server-service`
  - Selector: `app: server`
  - Ports:
    - name: grpc, port: 50051, targetPort: 50051
    - name: http, port: 8080, targetPort: 8080
  - Type: ClusterIP

**T3.3: Create Server HPA**
- **File:** `deployments/kustomize/base/server-hpa.yaml`
- **Content:**
  - HPA: `server-hpa`
  - ScaleTargetRef: Deployment/server
  - minReplicas: 2, maxReplicas: 10
  - Metrics: CPU 70%, Memory 80%

**T3.4: Create Server PDB**
- **File:** `deployments/kustomize/base/server-pdb.yaml`
- **Content:**
  - PDB: `server-pdb`
  - minAvailable: 1
  - Selector: `app: server`

#### Task 4: Create Client Deployment Resources

**T4.1: Create Client Deployment**
- **File:** `deployments/kustomize/base/client-deployment.yaml`
- **Content:**
  - Deployment: `client`, 2 replicas
  - Labels: `app: client`, `component: api-gateway`
  - Container: `zercle-go-template-client:latest`
  - Ports: 8080 (http)
  - EnvFrom: configmap, secret
  - Resources: requests 100m/128Mi, limits 500m/512Mi
  - LivenessProbe: HTTP GET /health on 8080
  - ReadinessProbe: HTTP GET /ready on 8080
  - SecurityContext: runAsNonRoot, runAsUser 65534

**T4.2: Create Client Service**
- **File:** `deployments/kustomize/base/client-service.yaml`
- **Content:**
  - Service: `client-service`
  - Selector: `app: client`
  - Ports:
    - name: http, port: 8080, targetPort: 8080
  - Type: ClusterIP

**T4.3: Create Client HPA**
- **File:** `deployments/kustomize/base/client-hpa.yaml`
- **Content:**
  - HPA: `client-hpa`
  - ScaleTargetRef: Deployment/client
  - minReplicas: 2, maxReplicas: 10
  - Metrics: CPU 70%, Memory 80%

**T4.4: Create Client PDB**
- **File:** `deployments/kustomize/base/client-pdb.yaml`
- **Content:**
  - PDB: `client-pdb`
  - minAvailable: 1
  - Selector: `app: client`

#### Task 5: Update Migration Job

**T5.1: Update Migration Job Config**
- **File:** `deployments/kustomize/base/migration-job.yaml`
- **Changes:**
  - Update name: `migration-job`
  - Update image: `zercle-go-template-migrate:latest`
  - Update env vars to use direct keys from new ConfigMap/Secret:
    ```yaml
    env:
      - name: APP_DATABASE_HOST
        valueFrom:
          configMapKeyRef:
            name: app-config
            key: APP_DATABASE_HOST
      # ... other env vars matching config structure
    ```
  - Keep backoffLimit: 3, ttlSecondsAfterFinished: 300

#### Task 6: Update Ingress

**T6.1: Update Ingress to Point to Client**
- **File:** `deployments/kustomize/base/ingress.yaml`
- **Changes:**
  - Update backend service name: `client-service`
  - Keep port: 8080
  - Update host placeholder to: `api.example.com` (customizable via overlay)

#### Task 7: Update Base Kustomization

**T7.1: Rewrite Base Kustomization**
- **File:** `deployments/kustomize/base/kustomization.yaml`
- **Changes:**
  - Remove `configMapGenerator` and `secretGenerator` (using explicit files instead)
  - Update `images` list:
    ```yaml
    images:
      - name: zercle-go-template-server
        newName: zercle-go-template-server
        newTag: latest
      - name: zercle-go-template-client
        newName: zercle-go-template-client
        newTag: latest
      - name: zercle-go-template-migrate
        newName: zercle-go-template-migrate
        newTag: latest
    ```
  - Update `resources` list:
    ```yaml
    resources:
      - namespace.yaml
      - serviceaccount.yaml
      - postgres-statefulset.yaml
      - valkey-statefulset.yaml
      - configmap.yaml
      - secret.yaml
      - server-deployment.yaml
      - server-service.yaml
      - server-hpa.yaml
      - server-pdb.yaml
      - client-deployment.yaml
      - client-service.yaml
      - client-hpa.yaml
      - client-pdb.yaml
      - migration-job.yaml
      - ingress.yaml
    ```

#### Task 8: Create Overlay Patches

**T8.1: Development Server Patch**
- **File:** `deployments/kustomize/overlays/development/patches/server-deployment-patch.yaml`
- **Content:**
  - Replicas: 1
  - Resources: requests 50m/64Mi, limits 200m/256Mi

**T8.2: Development Client Patch**
- **File:** `deployments/kustomize/overlays/development/patches/client-deployment-patch.yaml`
- **Content:**
  - Replicas: 1
  - Resources: requests 50m/64Mi, limits 200m/256Mi

**T8.3: Production Server Patch**
- **File:** `deployments/kustomize/overlays/production/patches/server-deployment-patch.yaml`
- **Content:**
  - Replicas: 3
  - Resources: requests 200m/256Mi, limits 1000m/1Gi

**T8.4: Production Client Patch**
- **File:** `deployments/kustomize/overlays/production/patches/client-deployment-patch.yaml`
- **Content:**
  - Replicas: 3
  - Resources: requests 200m/256Mi, limits 1000m/1Gi

#### Task 9: Update Overlay Kustomizations

**T9.1: Update Development Overlay**
- **File:** `deployments/kustomize/overlays/development/kustomization.yaml`
- **Changes:**
  - Update patches list:
    ```yaml
    patches:
      - path: patches/server-deployment-patch.yaml
        target:
          kind: Deployment
          name: server
      - path: patches/client-deployment-patch.yaml
        target:
          kind: Deployment
          name: client
    ```
  - Keep namespace: `zercle-go-template-dev`
  - Keep namePrefix: `dev-`
  - Update configMapGenerator (if keeping) to use correct env var names

**T9.2: Update Production Overlay**
- **File:** `deployments/kustomize/overlays/production/kustomization.yaml`
- **Changes:**
  - Update patches list (same structure as dev)
  - Keep namespace: `zercle-go-template-prod`
  - Keep namePrefix: `prod-`
  - Add SSL mode config: `APP_DATABASE_SSL_MODE: require`

#### Task 10: Cleanup Old Files

**T10.1: Remove Deprecated Files**
- Delete: `deployments/kustomize/base/deployment.yaml`
- Delete: `deployments/kustomize/base/service.yaml`
- Delete: `deployments/kustomize/base/hpa.yaml`
- Delete: `deployments/kustomize/base/pdb.yaml`
- Delete: `deployments/kustomize/overlays/development/patches/deployment-patch.yaml`
- Delete: `deployments/kustomize/overlays/production/patches/deployment-patch.yaml`

### N — Norms

**Naming Conventions:**
- Kubernetes resources: `kebab-case` (e.g., `server-deployment`, `postgres-service`)
- Labels: Use `app.kubernetes.io/name`, `app.kubernetes.io/component`, `app.kubernetes.io/part-of`
- Environment variables: `SCREAMING_SNAKE_CASE` with `APP_` prefix
- File names: descriptive, resource-type-first (e.g., `server-deployment.yaml` not `deployment-server.yaml`)

**Configuration Standards:**
- All env vars in ConfigMap/Secret must match viper's expectations with `APP_` prefix
- Config keys use underscore notation matching mapstructure tags (e.g., `APP_DATABASE_SSL_MODE` maps to `database.ssl_mode`)
- Sensitive values ONLY in Secrets
- Non-sensitive configuration in ConfigMap

**Security Standards:**
- All containers run as non-root (UID 65534)
- SecurityContext set on pod and container levels
- ServiceAccount with `automountServiceAccountToken: false`
- Resource limits set on all containers
- Readiness probes before accepting traffic

**Observability Standards:**
- Health checks on `/health` (liveness) and `/ready` (readiness)
- Structured logging (JSON in prod, text in dev)
- Labels for resource grouping and selection

**Resource Standards:**
- Requests: 50-100m CPU, 64-128Mi memory for dev
- Limits: 200-500m CPU, 256Mi-512Mi memory for dev
- Requests: 200m CPU, 256Mi memory for prod
- Limits: 1000m CPU, 1Gi memory for prod

### S — Safeguards

**Non-Negotiable Boundaries:**

1. **Environment Variable Contract:**
   - ALL env vars MUST use `APP_` prefix
   - ConfigMap and Secret keys MUST match what viper expects
   - Invalid: `SERVER_HOST`, `DB_HOST`, `CACHE_HOST`
   - Valid: `APP_SERVER_GRPC_HOST`, `APP_DATABASE_HOST`, `APP_VALKEY_HOST`

2. **Security Invariants:**
   - No container runs as root (UID must be 65534)
   - Secrets must NEVER be committed to git (use placeholder values)
   - Ingress must use TLS (even with placeholder certs)
   - Network policies should restrict pod-to-pod communication (future enhancement)

3. **Data Persistence:**
   - PostgreSQL MUST have a PVC for data durability
   - Valkey MUST have a PVC for data durability
   - PVC retention policy must be defined

4. **Availability Requirements:**
   - HPA minReplicas: 2 (prod), 1 (dev)
   - PDB minAvailable: 1
   - Readiness probes must pass before traffic routing

5. **Resource Limits:**
   - All containers MUST have CPU/memory limits
   - Requests must be ≤ 50% of limits

6. **Scope Exclusions (Not in this migration):**
   - No Helm charts (Kustomize only)
   - No Service Mesh (Istio/Linkerd)
   - No cert-manager integration (placeholder TLS)
   - No external secrets operator
   - No monitoring stack (Prometheus/Grafana) - assumes existing

**Validation Checklist:**
- [ ] All env vars start with `APP_`
- [ ] ConfigMap keys match viper config structure
- [ ] PostgreSQL StatefulSet has PVC
- [ ] Valkey StatefulSet has PVC
- [ ] Server deployment has both gRPC and HTTP ports
- [ ] Client deployment has HTTP port only
- [ ] Ingress points to client-service
- [ ] Migration job uses correct env vars
- [ ] All containers have resource limits
- [ ] All containers run as non-root
- [ ] Overlays correctly patch server and client separately

---

## Implementation Order

### Phase 1: Infrastructure (Tasks 1-2)
1. T1.1: Create PostgreSQL StatefulSet
2. T1.2: Create Valkey StatefulSet
3. T2.1: Create ConfigMap
4. T2.2: Create Secret

### Phase 2: Application Base (Tasks 3-6)
5. T3.1-T3.4: Create Server resources
6. T4.1-T4.4: Create Client resources
7. T5.1: Update Migration Job
8. T6.1: Update Ingress

### Phase 3: Kustomize Integration (Task 7)
9. T7.1: Update base kustomization.yaml

### Phase 4: Overlays (Tasks 8-9)
10. T8.1-T8.4: Create overlay patches
11. T9.1-T9.2: Update overlay kustomizations

### Phase 5: Cleanup (Task 10)
12. T10.1: Remove old files

---

## Files Summary

### New Files (18):
| File | Purpose |
|------|---------|
| `base/postgres-statefulset.yaml` | PostgreSQL StatefulSet |
| `base/valkey-statefulset.yaml` | Valkey StatefulSet |
| `base/configmap.yaml` | App configuration |
| `base/secret.yaml` | App secrets |
| `base/server-deployment.yaml` | Server deployment |
| `base/server-service.yaml` | Server service |
| `base/server-hpa.yaml` | Server HPA |
| `base/server-pdb.yaml` | Server PDB |
| `base/client-deployment.yaml` | Client deployment |
| `base/client-service.yaml` | Client service |
| `base/client-hpa.yaml` | Client HPA |
| `base/client-pdb.yaml` | Client PDB |
| `overlays/dev/patches/server-deployment-patch.yaml` | Dev server config |
| `overlays/dev/patches/client-deployment-patch.yaml` | Dev client config |
| `overlays/prod/patches/server-deployment-patch.yaml` | Prod server config |
| `overlays/prod/patches/client-deployment-patch.yaml` | Prod client config |

### Updated Files (5):
| File | Changes |
|------|---------|
| `base/kustomization.yaml` | New resources list, correct images |
| `base/migration-job.yaml` | Use correct env vars |
| `base/ingress.yaml` | Point to client-service |
| `overlays/dev/kustomization.yaml` | New patches |
| `overlays/prod/kustomization.yaml` | New patches |

### Deleted Files (6):
| File | Reason |
|------|--------|
| `base/deployment.yaml` | Replaced by server/client deployments |
| `base/service.yaml` | Replaced by server/client services |
| `base/hpa.yaml` | Replaced by server/client HPAs |
| `base/pdb.yaml` | Replaced by server/client PDBs |
| `overlays/dev/patches/deployment-patch.yaml` | Replaced by server/client patches |
| `overlays/prod/patches/deployment-patch.yaml` | Replaced by server/client patches |

---

## Environment Variable Reference

### ConfigMap Keys (Non-Sensitive):
```yaml
APP_SERVER_GRPC_HOST: "0.0.0.0"
APP_SERVER_GRPC_PORT: "50051"
APP_SERVER_HTTP_HOST: "0.0.0.0"
APP_SERVER_HTTP_PORT: "8080"
APP_DATABASE_HOST: "postgres-service"
APP_DATABASE_PORT: "5432"
APP_DATABASE_NAME: "zercle_chat"
APP_DATABASE_SSL_MODE: "disable"  # or "require" in prod
APP_DATABASE_MAX_CONNECTIONS: "25"
APP_DATABASE_MAX_IDLE_CONNS: "5"
APP_VALKEY_HOST: "valkey-service"
APP_VALKEY_PORT: "6379"
APP_VALKEY_DB: "0"
APP_LOGGING_LEVEL: "info"  # or "debug" in dev
APP_LOGGING_FORMAT: "json"  # or "text" in dev
```

### Secret Keys (Sensitive):
```yaml
APP_DATABASE_USER: "postgres"
APP_DATABASE_PASSWORD: "postgres"
APP_VALKEY_PASSWORD: ""  # empty if no auth
APP_AUTH_JWT_SECRET: "your-secret-key-change-in-production"
```

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Env var mismatch causes app crash | Medium | High | Use this plan's reference, test in minikube first |
| Migration job fails silently | Low | High | Add init container to wait for postgres, check logs |
| PVC data loss on StatefulSet delete | Low | High | Use Retain reclaim policy, document backup procedure |
| gRPC port not exposed correctly | Medium | Medium | Verify server-service has port 50051, test with grpcurl |
| HPA doesn't scale as expected | Low | Low | Monitor metrics, adjust thresholds in overlays |

---

## Testing Strategy

1. **Local Validation:**
   ```bash
   cd deployments/kustomize/overlays/development
   kustomize build . | kubectl apply --dry-run=client -f -
   ```

2. **Local Cluster (minikube/kind):**
   ```bash
   kustomize build overlays/development | kubectl apply -f -
   kubectl wait --for=condition=ready pod -l app=client --timeout=60s
   kubectl port-forward svc/client-service 8080:8080
   curl http://localhost:8080/health
   curl http://localhost:8080/ready
   ```

3. **Production Dry-Run:**
   ```bash
   kustomize build overlays/production | kubectl diff -f -
   ```
