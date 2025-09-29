# Quick Start Guide - Web UI

## 🚀 Get Started in 5 Minutes

### 1. Configure API Credentials

Create a `.env` file in the application directory:

```bash
# Slide API Configuration
SLIDE_API_URL=https://api.slide.tech
SLIDE_API_KEY=your_slide_api_key_here

# ConnectWise API Configuration
CONNECTWISE_API_URL=https://your-instance.connectwisedev.com/v4_6_release/apis/3.0
CONNECTWISE_COMPANY_ID=your_company_id
CONNECTWISE_PUBLIC_KEY=your_public_key
CONNECTWISE_PRIVATE_KEY=your_private_key
CONNECTWISE_CLIENT_ID=your_client_id
```

### 2. Start the Web UI

```bash
./slide-integrator.exe -web
```

You'll see:
```
Starting Slide-ConnectWise Integration with Web UI...
Alert monitor started in background
Web UI available at http://localhost:8080
```

### 3. Open Your Browser

Navigate to: **http://localhost:8080**

### 4. Configure Client Mappings

1. Click on the **🗺️ Client Mappings** tab
2. You'll see all your Slide clients listed
3. For each unmapped client:
   - Click **➕ Map**
   - Select the corresponding ConnectWise company from the dropdown
   - Click **Save Mapping**

**Pro Tip**: Click **🤖 Auto-Map Clients** to automatically map clients with matching names!

### 5. Configure Ticketing

1. Click on the **🎫 Ticketing Config** tab
2. Fill in the form:
   - **Board**: Select your service board
   - **Status**: Choose default status (e.g., "New")
   - **Priority**: Set priority level (e.g., "Medium")
   - **Type**: Select ticket type (e.g., "Issue")
3. Customize templates if needed:
   - Use variables like `{{client_name}}`, `{{alert_type}}`, `{{device_name}}`
   - Click **👁️ Preview** to see how it looks
4. Click **💾 Save Configuration**

### 6. Monitor Your Integration

Switch to the **📊 Dashboard** tab to see:
- Number of unresolved alerts
- Mapped clients count
- Open tickets

The **🚨 Alerts** tab shows all current alerts and their status.

The **📋 Tickets** tab displays the alert-to-ticket mappings.

---

## 🎯 Common Tasks

### View All Alerts
1. Go to **🚨 Alerts** tab
2. Use the search box to filter
3. Toggle "Show unresolved only" as needed

### Manually Close an Alert
1. Go to **🚨 Alerts** tab
2. Find the alert
3. Click **✓ Close**

### Delete a Mapping
1. Go to **🗺️ Client Mappings** tab
2. Find the mapped client
3. Click **🗑️ Delete**

### Change Port
If port 8080 is in use:
```bash
./slide-integrator.exe -web -port 3000
```

---

## 🔧 Troubleshooting

**Problem**: Web UI won't start
- Check if `.env` file exists with correct credentials
- Ensure port is not already in use (try different port with `-port` flag)

**Problem**: No clients showing up
- Verify Slide API credentials are correct
- Check application logs for API errors

**Problem**: Tickets not being created
- Ensure client mappings exist
- Verify ticketing configuration is saved
- Check ConnectWise API credentials

**Problem**: Can't access web UI remotely
- By default, the server only listens on localhost
- For remote access, consider setting up a reverse proxy (nginx, Apache)
- Or use SSH port forwarding: `ssh -L 8080:localhost:8080 user@server`

---

## 💡 Tips

1. **Auto-refresh**: Dashboard automatically refreshes every 30 seconds
2. **Template Variables**: Preview your templates before saving
3. **Search**: Use the search boxes to quickly find clients or alerts
4. **Background Service**: The alert monitor runs automatically in web UI mode
5. **Legacy CLI**: All CLI commands still work if you need them

---

## 📚 Next Steps

- Review the full [README.md](README.md) for advanced features
- Set up the service to run on system startup
- Consider creating backups of the database file
- Monitor logs for any API errors or issues

**Need help?** Check the logs in the terminal where you started the application.