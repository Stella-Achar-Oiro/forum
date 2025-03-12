// index.js - Global module imports and exports
// This file helps ensure all modules are loaded correctly

// Import core modules
import Component from './core/Component.js';
import Router from './core/Router.js';
import Store from './core/Store.js';
import ComponentRegistry from './core/ComponentRegistry.js';

// Import auth components and services
import AuthService from './services/AuthService.js';
import authGuard from './components/Auth/AuthGuard.js';
import LoginComponent from './components/Auth/LoginComponent.js';
import RegisterComponent from './components/Auth/RegisterComponent.js';

// Import post components
import PostsComponent from './components/Posts/PostsComponent.js';
import PostDetailComponent from './components/Posts/PostDetailComponent.js';
import CreatePostComponent from './components/Posts/CreatePostComponent.js';

// Make core modules available globally to handle module loading issues
window.ForumCore = {
    Component,
    Router,
    Store,
    ComponentRegistry,
    AuthService,
    authGuard
};

// Export all modules
export {
    Component,
    Router,
    Store,
    ComponentRegistry,
    AuthService,
    authGuard,
    LoginComponent,
    RegisterComponent,
    PostsComponent,
    PostDetailComponent,
    CreatePostComponent
};

console.log('Core modules loaded and exported globally via ForumCore object'); 