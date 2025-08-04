# FM-Dash

<div align="center">

![FM-Dash Logo](https://img.shields.io/badge/FM--Dash-Football%20Manager%20Data%20Analysis-blue?style=for-the-badge)

**Your ultimate Football Manager companion for finding the perfect players**

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![Vue Version](https://img.shields.io/badge/Vue.js-3.0+-4FC08D?style=flat-square&logo=vue.js)](https://vuejs.org)
[![Quasar Version](https://img.shields.io/badge/Quasar-2.0+-1976D2?style=flat-square&logo=quasar)](https://quasar.dev)
[![License](https://img.shields.io/badge/License-CC%20BY--NC--SA%204.0-lightgrey.svg?style=flat-square)](https://creativecommons.org/licenses/by-nc-sa/4.0/)

[![CI/CD](https://github.com/LiamHardman/fm-dash/actions/workflows/code-quality.yml/badge.svg)](https://github.com/LiamHardman/fm-dash/actions/workflows/code-quality.yml)
[![Release](https://github.com/LiamHardman/fm-dash/actions/workflows/release.yml/badge.svg)](https://github.com/LiamHardman/fm-dash/actions/workflows/release.yml)
[![Docker](https://github.com/LiamHardman/fm-dash/actions/workflows/deploy.yml/badge.svg)](https://github.com/LiamHardman/fm-dash/actions/workflows/deploy.yml)
[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/LiamHardman/fm-dash?sort=semver&style=flat-square)](https://github.com/LiamHardman/fm-dash/releases)
[![GitHub issues](https://img.shields.io/github/issues/LiamHardman/fm-dash?style=flat-square)](https://github.com/LiamHardman/fm-dash/issues)
[![GitHub stars](https://img.shields.io/github/stars/LiamHardman/fm-dash?style=flat-square)](https://github.com/LiamHardman/fm-dash/stargazers)

</div>

## What is FM-Dash?

FM-Dash transforms your Football Manager HTML exports into a powerful scouting tool. Upload your squad data and discover hidden gems, find perfect replacements, and build the ultimate team with intelligent analysis and beautiful visualizations.

## ✨ What You Can Do

### 🔍 **Find Hidden Gems**
- **Bargain Hunter**: Discover undervalued players who punch above their weight
- **Wonderkids Discovery**: Unearth the next generation of superstars
- **Free Agents**: Find quality players available on free transfers
- **Upgrade Finder**: Get recommendations to improve your current squad

### 📊 **Analyze Like a Pro**
- **Smart Filtering**: Search by position, nationality, age, league, and more
- **Performance Ratings**: FIFA-style ratings and percentile rankings
- **Detailed Profiles**: Deep dive into every player's attributes
- **Pitch Visualization**: See your team formation and tactical insights

### 🎯 **Build Your Dream Team**
- **Wishlist System**: Save and track your favorite players
- **Team Logos**: Beautiful visual integration with club branding
- **League Overviews**: Comprehensive statistics and insights
- **Player Photos**: Real player faces for immersive experience

### ⚡ **Lightning Fast Performance**
- **Handle Large Datasets**: Process 50MB+ files with ease
- **Smooth Scrolling**: Navigate thousands of players effortlessly
- **Instant Search**: Find players in milliseconds

## 🚀 Quick Start

### Try It Online
Visit our live demo to see FM-Dash in action!

### Run Locally

1. **Get the Code**
   ```bash
   git clone https://github.com/LiamHardman/fm-dash.git
   cd fm-dash
   ```

2. **Start Development Mode**
   ```bash
   # Install dependencies and start both frontend and backend
   ./scripts/setup-dev.sh
   npm run serve  # Backend API
   npm run dev    # Frontend (in another terminal)
   ```

3. **Access the App**
   - Frontend: http://localhost:3000
   - API: http://localhost:8091

### Docker Quick Start
```bash
docker build -t fm-dash .
docker run -p 8080:8080 fm-dash
```
Then visit http://localhost:8080

## 🎮 Key Features Explained

### Bargain Hunter
Find players who offer exceptional value for money. Perfect for clubs with limited budgets who need to maximize their transfer spending.

### Wonderkids Discovery
Identify the next generation of superstars. Filter by age and potential to find young players who could become world-class.

### Upgrade Finder
Get specific recommendations for improving your current squad. Compare players side-by-side and find the perfect replacements.

### Interactive Pitch Display
Visualize your team formation and see how players fit together tactically. Great for understanding team chemistry and tactical fit.

## 🤝 Contributing

We love contributions! Whether you're a developer, designer, or Football Manager enthusiast, there are many ways to help:

- **Report Bugs**: Found an issue? Let us know!
- **Suggest Features**: Have ideas for improvements?
- **Improve Documentation**: Help make things clearer
- **Code Contributions**: Submit pull requests

See our [Contributing Guide](CONTRIBUTING.md) for details.

## 📚 Documentation

- **[User Guide](docs/USER_GUIDE.md)** - Complete guide to using FM-Dash
- **[API Reference](docs/API.md)** - For developers and integrations
- **[Configuration](docs/CONFIGURATION.md)** - Setup and customization options
- **[Troubleshooting](docs/TROUBLESHOOTING.md)** - Common issues and solutions
- **[Architecture](docs/ARCHITECTURE.md)** - Technical architecture and design decisions
- **[Frontend Performance](docs/FRONTEND_PERFORMANCE.md)** - Performance optimization guide
- **[Contributing](CONTRIBUTING.md)** - How to contribute to the project

## 🆘 Support

Need help? Here's where to find it:

- **Documentation**: Check our comprehensive guides
- **Issues**: Report bugs or request features on GitHub
- **Discussions**: Join the community conversation
- **Changelog**: See what's new in recent updates

## 🙏 Credits

Special thanks to:
- **sortitoutsi** and **Footygamer** for providing club logos and player faces
- The Football Manager community for feedback and suggestions
- All contributors who help make FM-Dash better

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

**Ready to revolutionize your Football Manager scouting?** 🚀

For technical details, architecture information, and development setup, see [TECHNICAL_DETAILS.md](TECHNICAL_DETAILS.md).


