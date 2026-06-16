# VibeCode Rescue Blog Site

A personal blog about vibecoding and rescuing vibecoded codebases with the intrane.fr automaintainer proposal.

## Overview

This site contains 50 blog posts covering:
- Vibecoding principles and practices
- Automaintainer features and implementation
- Team culture, leadership, and collaboration
- Technical topics like refactoring, testing, security
- References to the intrane.fr automaintainer proposal

## Setup

Run the setup script to create the site:

```bash
./scripts/setup-vibecode-rescue.sh
```

Or manually:

```bash
# Create site
./sl-cli site create --name "VibeCode Rescue" --slug "vibecode-rescue"
./sl-cli site version create vibecode-rescue --version "v1"

# Copy files
cp -r sites/vibecode-rescue/v1/* ~/.superlandings/sites/vibecode-rescue/v1/
```

## Blog Posts

The site includes 50 blog posts organized into categories:

### Vibecoding Topics
- Introduction to Vibecoding
- Core Principles of Vibecoding
- Rescuing Vibecoded Codebases
- Team Culture
- Leadership
- Communication
- Decision Making
- Feedback
- Remote Collaboration
- Time Management
- Documentation
- Code Quality
- Continuous Improvement
- Burnout Prevention
- Legacy Code Empathy
- Mentoring
- Failure Learning
- Code Review
- Refactoring
- Psychological Safety
- Celebration
- Hiring
- Onboarding
- Knowledge Sharing
- Collaboration Tools

### Automaintainer Topics
- Automaintainer Overview
- The Future of Automaintainer
- Automaintainer Architecture
- Security Best Practices
- Integration
- Performance
- Cost-Benefit Analysis
- Metrics and Reporting
- Observability
- Migration
- Disaster Recovery
- Privacy
- Best Practices
- Scaling
- AI in Automaintainer
- Security
- Testing
- Customization
- Rules
- Roadmap
- Multitenancy
- Open Source
- Compliance
- Technical Debt Strategies

## Technical Details

- **Framework**: SuperLandings Go
- **Styling**: Tailwind CSS (CDN)
- **Templating**: Go templates with includes
- **Content**: Markdown blog posts
- **Storage**: File system based
- **Versioning**: v1

## intrane.fr References

Several posts reference the intrane.fr automaintainer proposal:
- Automaintainer Overview
- The Future of Automaintainer  
- Integrating Automaintainer

## License

MIT