# MongoDB Setup Guide for Diabetes Assistant

This application requires MongoDB for storing user data. You have two options:

## Option 1: Local MongoDB Installation

### macOS

1. Install via Homebrew:
   ```
   brew tap mongodb/brew
   brew install mongodb-community
   ```

2. Start MongoDB:
   ```
   brew services start mongodb-community
   ```

3. Verify it's running:
   ```
   brew services list
   ```

### Ubuntu/Debian

1. Import the MongoDB public GPG key:
   ```
   wget -qO - https://www.mongodb.org/static/pgp/server-6.0.asc | sudo apt-key add -
   ```

2. Create a list file for MongoDB:
   ```
   echo "deb [ arch=amd64,arm64 ] https://repo.mongodb.org/apt/ubuntu focal/mongodb-org/6.0 multiverse" | sudo tee /etc/apt/sources.list.d/mongodb-org-6.0.list
   ```

3. Update the package database:
   ```
   sudo apt-get update
   ```

4. Install MongoDB:
   ```
   sudo apt-get install -y mongodb-org
   ```

5. Start MongoDB:
   ```
   sudo systemctl start mongod
   ```

6. Enable MongoDB to start on system reboot:
   ```
   sudo systemctl enable mongod
   ```

7. Verify it's running:
   ```
   sudo systemctl status mongod
   ```

### Windows

1. Download the MongoDB Community Server from [MongoDB Download Center](https://www.mongodb.com/try/download/community)
2. Run the installer and follow the instructions
3. Choose "Complete" setup
4. Select "Run service as Network Service user"
5. Uncheck "Install MongoDB Compass" if you don't need the GUI tool
6. Complete the installation

## Option 2: MongoDB Atlas (Cloud Database)

If you prefer not to install MongoDB locally, you can use MongoDB Atlas, which offers a free tier that's sufficient for this application.

1. Create a free account at [MongoDB Atlas](https://www.mongodb.com/cloud/atlas/register)
2. Create a new cluster (choose the free tier option)
3. Set up a database user with password authentication
4. Configure network access (IP access list)
   - Add your current IP address
   - Or add `0.0.0.0/0` to allow access from anywhere (less secure, but simpler)
5. Get your connection string, which will look similar to:
   ```
   mongodb+srv://username:password@cluster0.mongodb.net/diabetes_assistant
   ```
6. Update your `.env` file with this connection string

## Troubleshooting MongoDB Connection Issues

### Connection Refused

If you see "connection refused" errors:

1. Check if MongoDB is running:
   - macOS: `brew services list`
   - Linux: `sudo systemctl status mongod`
   - Windows: Check Services in Task Manager

2. Verify the connection string in your `.env` file:
   - For local MongoDB, it should be `mongodb://localhost:27017/diabetes_assistant`
   - For MongoDB Atlas, ensure you've replaced `username:password` with your actual credentials

3. Check network configuration:
   - If using MongoDB Atlas, ensure your IP is in the access list
   - If behind a firewall, ensure the MongoDB port (27017 for local, outbound HTTPS for Atlas) is open

### Authentication Failed

If you see authentication errors:

1. Double-check username and password in your connection string
2. For MongoDB Atlas, verify that you created a database user correctly
3. Check that your Atlas user has the correct permissions (Atlas admin or readWrite on your database)

### Other Issues

For other connection issues:

1. Try connecting with MongoDB Compass to verify your connection string works
2. Check MongoDB logs:
   - macOS: `/usr/local/var/log/mongodb/mongo.log`
   - Linux: `/var/log/mongodb/mongod.log`
   - Windows: Check the location specified in your MongoDB configuration 