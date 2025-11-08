## **Problem Statement**

You're building a sync service for a developer platform that hosts hackathons and project submissions.

## **Challenge**

The platform uses **Postgres as the source of truth** for all data, but relies on **Elasticsearch for search functionality**. When data changes in Postgres, those changes must be reflected in Elasticsearch. Your task is to design and build a system that keeps both datastores in sync.

## **Data**

Your system needs to handle three main entities:

- **Users** - Developer profiles (username, email, skills, college, etc.)
- **Hackathons** - Coding events (name, dates, location, tracks, etc.)
- **Projects** - Hackathon submissions (name, description, team members, hackathon, etc.)

These entities have relationships:

- Users participate in Hackathons
- Projects belong to Hackathons and are created by Users
- Think about how changes cascade (e.g., what happens when a user updates their profile?)

You can choose to construct the data model for this, however you see most appropriate.

## **Core Requirements**

1. **Capture database changes** - Detect when records are created, updated, or deleted in Postgres

2. **Sync to Elasticsearch** - Apply those changes to corresponding Elasticsearch indexes

3. **Handle failures gracefully** - Network issues, Elasticsearch downtime, malformed data

4. **Maintain data consistency** - Ensure Elasticsearch accurately reflects Postgres state

5. **Provide visibility** - Some way to monitor sync health and debug issues

## **Questions to Consider**

- How will you capture changes from Postgres?
- What happens if Elasticsearch is temporarily unavailable?
- How do you ensure changes are applied in the correct order?
- How would you recover from a complete Elasticsearch data loss?
- What are the trade-offs in your design?

## **Tech Stack Reference**

We use:

- **Language:** Go
- **Database:** Postgres with GORM
- **Search:** Elasticsearch
- **Architecture:** HTTP API with CRON-triggered sync jobs