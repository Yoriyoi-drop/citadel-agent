# Citadel Agent - Troubleshooting Guide

## Common Issues and Solutions

### Issue: "Can't connect: Resource temporarily unavailable"
**Symptoms:**
- When starting the service, you see: "Can't connect: Resource temporarily unavailable"
- The application fails to start properly

**Causes and Solutions:**

1. **Port Already in Use**
   - **Cause**: Another process is using the same port (typically 5001 for API)
   - **Solution**: 
     ```bash
     # Check what's using port 5001
     lsof -i :5001
     # Kill the process if it's not needed
     kill -9 <PID>
     # Or change the port in your .env file:
     echo "SERVER_PORT=5002" >> .env
     ```

2. **Database Not Ready**
   - **Cause**: PostgreSQL takes time to initialize, and the API tries to connect before it's ready
   - **Solution**: The enhanced docker-compose.yml now has proper health checks, but if you're still having issues:
     ```bash
     # Bring down services completely
     docker-compose -f docker/docker-compose.yml down
     # Remove any orphaned containers
     docker container prune
     # Restart with fresh setup
     ./scripts/start.sh
     ```

3. **Docker Resources Insufficient**
   - **Cause**: Docker is running out of memory/disk space
   - **Solution**:
     ```bash
     # Check Docker disk usage
     docker system df
     # Clean up unused resources
     docker system prune -a --volumes
     # Increase Docker memory allocation in Docker Desktop settings
     ```

### Issue: API Service Crashes Immediately
**Symptoms:**
- API service crashes right after starting
- Check logs with: `docker-compose -f docker/docker-compose.yml logs api`

**Common Causes:**
1. **Environment Variables Missing**
   - Solution: Ensure you have a `.env` file with required variables:
     ```
     DB_USER=postgres
     DB_PASSWORD=postgres
     DB_NAME=citadel_agent
     JWT_SECRET=your-super-secret-jwt-key-here-change-in-production
     SERVER_PORT=5001
     ```

2. **Database Connection Issues**
   - Check if PostgreSQL is running and healthy
   - Look for database migration issues in logs

### Issue: Worker/Scheduler Services Don't Start
**Symptoms:**
- Worker or scheduler services don't start or crash
- May be related to database setup

**Solution:**
1. Verify the database is fully initialized
2. Check that migrations are run properly
3. Make sure API service starts first (in our setup, dependencies are handled)

## Startup Scripts Usage

### Quick Start
```bash
# Start the entire stack
./scripts/start.sh

# Check service status
./scripts/status.sh

# Stop the stack
./scripts/stop.sh
```

### Development Mode
```bash
# For development with hot reloading
docker-compose -f docker/docker-compose.dev.yml up --build
```

## Health Check Endpoints

### API Service
- Health: `http://localhost:5001/health`
- Should return: `{"status":"OK","time":<timestamp>}`

### Database Connections
- PostgreSQL: Connects to localhost:5432 (mapped from container)
- Redis: Connects to localhost:6379 (mapped from container)

## Database Migration

If you get database-related errors, you may need to run migrations:

```bash
# Check if migrations are needed by looking at logs
docker-compose -f docker/docker-compose.yml logs api | grep -i migration
```

## Debugging Steps

1. **Check Docker Status:**
   ```bash
   docker ps
   docker-compose -f docker/docker-compose.yml ps
   ```

2. **Check Service Logs:**
   ```bash
   # View all logs
   docker-compose -f docker/docker-compose.yml logs
   
   # Follow logs in real-time
   docker-compose -f docker/docker-compose.yml logs -f
   
   # Check specific service
   docker-compose -f docker/docker-compose.yml logs api
   ```

3. **Verify Environment:**
   ```bash
   # Check if environment variables are loaded
   docker-compose -f docker/docker-compose.yml config
   ```

4. **Manual Health Checks:**
   ```bash
   # Test connectivity to database
   docker-compose -f docker/docker-compose.yml exec postgres psql -U postgres -c "\dt"
   
   # Test Redis
   docker-compose -f docker/docker-compose.yml exec redis redis-cli ping
   ```

## Performance Considerations

- **Memory Usage**: The full stack typically uses 1-2 GB of RAM
- **Disk Space**: Requires ~2GB of free space for Docker images and volumes
- **Startup Time**: Allow 1-2 minutes for all services to fully start

## Common Commands

```bash
# View all containers
docker ps

# View logs for a specific service
docker-compose -f docker/docker-compose.yml logs api

# Execute commands inside a container  
docker-compose -f docker/docker-compose.yml exec api sh

# Scale specific services
docker-compose -f docker/docker-compose.yml up -d --scale worker=2

# Clean up everything (removes data!)
docker-compose -f docker/docker-compose.yml down -v
```

## When Everything Else Fails

Sometimes a complete reset is needed:

```bash
# Stop all services
./scripts/stop.sh

# Remove all containers, networks, and volumes
docker-compose -f docker/docker-compose.yml down -v

# Clear Docker system
docker system prune -a --volumes

# Restart fresh
./scripts/start.sh
```

Remember: `-v` flag removes volumes (persistent data), so don't use it if you want to keep your workflows and data.