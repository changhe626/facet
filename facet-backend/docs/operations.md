# Daily Operations Manual

This document provides the core daily operations commands and Standard Operating Procedures (SOP) for the `facet-backend` project, designed to ensure safe and efficient deployment and maintenance.

**Basic Principles**: All commands should be executed from the project root directory (i.e., the directory where the `docker-compose.yml` file is located). Always use the new generation `docker compose` command (with a space).

---

## 1. Status Monitoring

Before and after performing any operations, always check the status first.

#### View All Service Statuses

Check whether all services (`backend`, `frontend`, `postgres`, `redis`) are in a healthy `Up` or `running` state.

```bash
docker compose ps
```

#### View All Service Logs (Aggregated)

View the log output of all containers at once, suitable for a quick overview of global activity.

```bash
docker compose logs
```

#### Real-Time Tracking of Specific Service Logs (Most Common)

This is the preferred command for troubleshooting. It outputs the logs of the specified service in real time. Press `Ctrl+C` to exit.

```bash
# View backend logs in real time
docker compose logs -f backend

# View frontend logs in real time
docker compose logs -f frontend
```

---

## 2. Core Deployment Process

#### Scenario 1: Full Deployment (Safest, Most Recommended)

When there are any code changes (whether frontend or backend), this standard process is recommended. It ensures all services use the latest images and the environment is clean.

```bash
# Step 1: Stop and remove all old containers and networks
# (Note: This will not delete database and Redis data volumes)
docker compose down

# Step 2: Rebuild images for all services and start new containers in the background
docker compose up --build -d

# Step 3: Verify that all services have started successfully
docker compose ps
```
> **Note**: This process will not automatically delete old backend images. For information on cleaning up these old images, see section "5. System Cleanup and Maintenance".

#### Scenario 2: Deploy Only Backend Service (Fast Iteration)

When only the backend code (`facet-backend`) has changed, use this process to avoid unnecessary frontend service restarts and speed up deployment.

```bash
# Step 1: Stop the currently running backend container
docker compose stop backend

# Step 2: Force remove the stopped backend container
docker compose rm -f backend

# Step 3: Only rebuild the backend service image
docker compose build backend

# Step 4: Only start the backend service without affecting its dependencies (postgres, redis)
docker compose up -d --no-deps backend

# Step 5: Verify backend service status and logs
docker compose ps
# View logs
docker compose logs -f backend
```

#### Scenario 3: Deploy Only Frontend Service (Fast Iteration)

When only the frontend code (`facet-front`) has changed, use this process.

```bash
# Step 1: Stop the currently running frontend container
docker compose stop frontend

# Step 2: Force remove the stopped frontend container
docker compose rm -f frontend

# Step 3: Only rebuild the frontend service image
docker compose build frontend

# Step 4: Only start the frontend service without affecting the backend service
docker compose up -d --no-deps frontend

# Step 5: Verify frontend service status
docker compose ps
# View logs
docker compose logs -f frontend
```

---

## 3. Daily Maintenance Operations

#### Restart All Services

If the entire system is experiencing issues or you need to apply configuration changes to all services at once, use this command. It will restart all services in order.

```bash
docker compose restart
```

#### Restart a Single Service

If only a specific service has an issue (e.g., the backend service needs to reload a configuration), use this command for a quick restart without affecting other services.

```bash
# Restart the backend service (most common)
docker compose restart backend

# Restart the frontend service
docker compose restart frontend
```

#### Enter a Container for Troubleshooting

For certain complex issues (e.g., needing to check files, network, processes inside a container), you need to enter the container's shell environment.

```bash
# Enter the backend container's shell environment
docker compose exec backend sh

# Once inside, you can execute Linux commands such as ls -l, cat logs/app.log, top, etc.
```

---

## 4. Destructive Operations

**Warning: The following commands will permanently delete data. Please confirm and back up before execution.**

#### Complete Cleanup of Everything

This command stops and removes all containers, networks, **all related images**, and **all data volumes**. After execution, all data in PostgreSQL and Redis will be **permanently lost**.

This is typically only used when a completely fresh, clean development environment is needed.

```bash
# --rmi all: Delete all related images
# -v: Delete all data volumes defined in docker-compose.yml
docker compose down --rmi all -v
```

## 5. System Cleanup and Maintenance

#### Why Cleanup Is Needed?

When you run the `docker compose build` command, Docker creates a new image for your code and tags it with something like `facet-backend-backend:latest`. If an old image with the same name and tag already exists, Docker does not delete it; instead, it "moves" the tag from the old image to the new one. At this point, the old image becomes an untagged "**dangling image**".

As deployments continue, these dangling images accumulate and eventually fill up the server's disk space. Therefore, **treating "daily deployment" and "periodic cleanup" as two separate processes** is the best practice.

#### Clean Up Unused Docker Images (Recommended Routine Operation)

This command **safely** removes all dangling images that are not referenced by any container. It is strongly recommended to perform this maintenance operation regularly (e.g., weekly or monthly).

```bash
docker image prune
```

> The system will prompt you to confirm the deletion. Enter `y` and press Enter.

#### Clean Up Unused Docker Resources (More Thorough)

If you want a more comprehensive cleanup, including all stopped containers, all networks not used by any container, and all dangling images, use the following command:

```bash
docker system prune
```

> **Warning**: This command will delete all containers that are **not running**. Before executing, make sure there are no stopped containers you wish to keep.
