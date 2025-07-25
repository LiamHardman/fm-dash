#!/usr/bin/env node

/**
 * Bundle Analysis Script
 * Analyzes the built bundle to identify optimization opportunities
 */

import { readFileSync, existsSync } from 'fs'
import { join } from 'path'
import { execSync } from 'child_process'

const BUNDLE_SIZE_LIMITS = {
  critical: 300 * 1024, // 300KB for critical chunks
  warning: 500 * 1024,  // 500KB warning threshold
  error: 1000 * 1024    // 1MB error threshold
}

class BundleAnalyzer {
  constructor() {
    this.distPath = join(process.cwd(), 'dist')
    this.results = {
      chunks: [],
      assets: [],
      warnings: [],
      recommendations: []
    }
  }

  async analyze() {
    console.log('🔍 Starting bundle analysis...\n')

    if (!existsSync(this.distPath)) {
      console.error('❌ Build directory not found. Run "npm run build" first.')
      process.exit(1)
    }

    this.analyzeChunks()
    this.analyzeAssets()
    this.generateRecommendations()
    this.printReport()
  }

  analyzeChunks() {
    console.log('📦 Analyzing JavaScript chunks...')
    
    try {
      const jsFiles = execSync(`find ${this.distPath} -name "*.js" -type f`, { encoding: 'utf8' })
        .trim()
        .split('\n')
        .filter(Boolean)

      for (const filePath of jsFiles) {
        const stats = this.getFileStats(filePath)
        const fileName = filePath.split('/').pop()
        
        this.results.chunks.push({
          name: fileName,
          path: filePath,
          size: stats.size,
          sizeKB: Math.round(stats.size / 1024),
          type: this.getChunkType(fileName),
          isEntry: fileName.includes('index-') || fileName.includes('main-'),
          isVendor: fileName.includes('vendor-') || fileName.includes('chunk-'),
          isPage: fileName.includes('Page-') || fileName.includes('page-')
        })

        // Check for size warnings
        if (stats.size > BUNDLE_SIZE_LIMITS.error) {
          this.results.warnings.push({
            type: 'error',
            message: `Chunk ${fileName} is too large (${Math.round(stats.size / 1024)}KB)`,
            recommendation: 'Consider splitting this chunk further or lazy loading components'
          })
        } else if (stats.size > BUNDLE_SIZE_LIMITS.warning) {
          this.results.warnings.push({
            type: 'warning',
            message: `Chunk ${fileName} is large (${Math.round(stats.size / 1024)}KB)`,
            recommendation: 'Consider optimizing or splitting this chunk'
          })
        }
      }

      // Sort chunks by size
      this.results.chunks.sort((a, b) => b.size - a.size)
      
    } catch (error) {
      console.error('Error analyzing chunks:', error.message)
    }
  }

  analyzeAssets() {
    console.log('🎨 Analyzing CSS and other assets...')
    
    try {
      const cssFiles = execSync(`find ${this.distPath} -name "*.css" -type f`, { encoding: 'utf8' })
        .trim()
        .split('\n')
        .filter(Boolean)

      for (const filePath of cssFiles) {
        const stats = this.getFileStats(filePath)
        const fileName = filePath.split('/').pop()
        
        this.results.assets.push({
          name: fileName,
          path: filePath,
          size: stats.size,
          sizeKB: Math.round(stats.size / 1024),
          type: 'css'
        })
      }

      // Analyze other assets (images, fonts, etc.)
      const otherAssets = execSync(`find ${this.distPath} -type f ! -name "*.js" ! -name "*.css" ! -name "*.html" ! -name "*.map"`, { encoding: 'utf8' })
        .trim()
        .split('\n')
        .filter(Boolean)

      for (const filePath of otherAssets) {
        const stats = this.getFileStats(filePath)
        const fileName = filePath.split('/').pop()
        const ext = fileName.split('.').pop()
        
        this.results.assets.push({
          name: fileName,
          path: filePath,
          size: stats.size,
          sizeKB: Math.round(stats.size / 1024),
          type: ext
        })
      }

      // Sort assets by size
      this.results.assets.sort((a, b) => b.size - a.size)
      
    } catch (error) {
      console.error('Error analyzing assets:', error.message)
    }
  }

  generateRecommendations() {
    console.log('💡 Generating optimization recommendations...')

    const totalJSSize = this.results.chunks.reduce((sum, chunk) => sum + chunk.size, 0)
    const totalCSSSize = this.results.assets
      .filter(asset => asset.type === 'css')
      .reduce((sum, asset) => sum + asset.size, 0)

    // Check for large vendor chunks
    const largeVendorChunks = this.results.chunks.filter(chunk => 
      chunk.isVendor && chunk.size > BUNDLE_SIZE_LIMITS.warning
    )

    if (largeVendorChunks.length > 0) {
      this.results.recommendations.push({
        type: 'optimization',
        title: 'Large Vendor Chunks Detected',
        description: `Found ${largeVendorChunks.length} large vendor chunk(s)`,
        action: 'Consider splitting vendor libraries into smaller, more focused chunks',
        chunks: largeVendorChunks.map(c => c.name)
      })
    }

    // Check for duplicate dependencies
    const vendorChunks = this.results.chunks.filter(chunk => chunk.isVendor)
    if (vendorChunks.length > 5) {
      this.results.recommendations.push({
        type: 'optimization',
        title: 'Many Vendor Chunks',
        description: `Found ${vendorChunks.length} vendor chunks`,
        action: 'Consider consolidating related vendor libraries to reduce HTTP requests'
      })
    }

    // Check total bundle size
    const totalSizeMB = (totalJSSize + totalCSSSize) / (1024 * 1024)
    if (totalSizeMB > 2) {
      this.results.recommendations.push({
        type: 'performance',
        title: 'Large Total Bundle Size',
        description: `Total bundle size is ${totalSizeMB.toFixed(2)}MB`,
        action: 'Consider implementing more aggressive code splitting and lazy loading'
      })
    }

    // Check for unused CSS
    const largeCSSFiles = this.results.assets.filter(asset => 
      asset.type === 'css' && asset.size > 50 * 1024
    )

    if (largeCSSFiles.length > 0) {
      this.results.recommendations.push({
        type: 'optimization',
        title: 'Large CSS Files',
        description: `Found ${largeCSSFiles.length} large CSS file(s)`,
        action: 'Consider using CSS purging or splitting CSS by route',
        files: largeCSSFiles.map(f => f.name)
      })
    }
  }

  printReport() {
    console.log('\n📊 Bundle Analysis Report')
    console.log('=' .repeat(50))

    // Summary
    const totalJSSize = this.results.chunks.reduce((sum, chunk) => sum + chunk.size, 0)
    const totalCSSSize = this.results.assets
      .filter(asset => asset.type === 'css')
      .reduce((sum, asset) => sum + asset.size, 0)
    const totalAssetSize = this.results.assets
      .filter(asset => asset.type !== 'css')
      .reduce((sum, asset) => sum + asset.size, 0)

    console.log('\n📈 Summary:')
    console.log(`  JavaScript: ${Math.round(totalJSSize / 1024)}KB (${this.results.chunks.length} files)`)
    console.log(`  CSS: ${Math.round(totalCSSSize / 1024)}KB`)
    console.log(`  Other Assets: ${Math.round(totalAssetSize / 1024)}KB`)
    console.log(`  Total: ${Math.round((totalJSSize + totalCSSSize + totalAssetSize) / 1024)}KB`)

    // Top 10 largest chunks
    console.log('\n🏆 Largest JavaScript Chunks:')
    this.results.chunks.slice(0, 10).forEach((chunk, index) => {
      const icon = chunk.isEntry ? '🚀' : chunk.isVendor ? '📦' : chunk.isPage ? '📄' : '🔧'
      console.log(`  ${index + 1}. ${icon} ${chunk.name} - ${chunk.sizeKB}KB`)
    })

    // Warnings
    if (this.results.warnings.length > 0) {
      console.log('\n⚠️  Warnings:')
      this.results.warnings.forEach(warning => {
        const icon = warning.type === 'error' ? '❌' : '⚠️'
        console.log(`  ${icon} ${warning.message}`)
        console.log(`     💡 ${warning.recommendation}`)
      })
    }

    // Recommendations
    if (this.results.recommendations.length > 0) {
      console.log('\n💡 Optimization Recommendations:')
      this.results.recommendations.forEach((rec, index) => {
        const icon = rec.type === 'performance' ? '⚡' : '🔧'
        console.log(`  ${index + 1}. ${icon} ${rec.title}`)
        console.log(`     ${rec.description}`)
        console.log(`     Action: ${rec.action}`)
        if (rec.chunks) {
          console.log(`     Affected: ${rec.chunks.join(', ')}`)
        }
        if (rec.files) {
          console.log(`     Files: ${rec.files.join(', ')}`)
        }
        console.log()
      })
    }

    // Success message
    if (this.results.warnings.length === 0) {
      console.log('\n✅ Bundle analysis complete - no critical issues found!')
    }

    console.log('\n🔗 For detailed analysis, run: npm run build:analyze')
    console.log('   This will generate an interactive bundle visualization.')
  }

  getFileStats(filePath) {
    try {
      const content = readFileSync(filePath)
      return {
        size: content.length
      }
    } catch (error) {
      return { size: 0 }
    }
  }

  getChunkType(fileName) {
    if (fileName.includes('vendor-')) return 'vendor'
    if (fileName.includes('page-') || fileName.includes('Page-')) return 'page'
    if (fileName.includes('component-')) return 'component'
    if (fileName.includes('shared-')) return 'shared'
    if (fileName.includes('index-') || fileName.includes('main-')) return 'entry'
    return 'chunk'
  }
}

// Run the analyzer
const analyzer = new BundleAnalyzer()
analyzer.analyze().catch(error => {
  console.error('Analysis failed:', error)
  process.exit(1)
})