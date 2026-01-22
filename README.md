# 🎸 Groupie Tracker

![Go Version](https://img.shields.io/badge/Go-1.25.0-blue?style=flat&logo=go)
![Fyne Version](https://img.shields.io/badge/Fyne-v2.7.2-orange?style=flat&logo=gui)
![Status](https://img.shields.io/badge/Status-Completed-success)

Groupie Tracker est une application de bureau performante développée en Go utilisant le framework graphique Fyne. Elle offre une interface ergonomique pour visualiser, rechercher et manipuler des données sur des artistes et groupes de musique via une API RESTful, tout en respectant les principes de conception de Shneiderman.

---

## 📑 Table des Matières
- Description
- Fonctionnalités
- Stack Technique
- Prérequis
- Installation et Démarrage
- Structure du Projet
- Auteurs

---

## 📋 Description

Ce projet étudiant (Ynov) dépasse le simple cadre de la visualisation de données JSON. Il propose une expérience utilisateur fluide permettant d'explorer l'univers musical, de géolocaliser des concerts et de gérer des données personnelles.

L'application récupère les données (artistes, lieux, dates, relations) depuis une API distante et permet également à l'utilisateur d'enrichir cette base de données localement.

---

## ✨ Fonctionnalités

### Recherche et Exploration Avancées
- Barre de recherche intelligente : Filtrage en temps réel par nom d'artiste, membre ou lieu.
- Filtres dynamiques :
  - Dates : Création du groupe et sortie du premier album (Range Selectors).
  - Membres : Sélection par nombre de membres (duo, trio, etc.).
  - Localisation : Filtrage par ville de concert.
- Tri : Ordonnancement par nom, date de création ou premier album.

### Géolocalisation et Cartographie
- OpenStreetMap Integration : Utilisation de l'API Nominatim pour convertir les lieux de concerts en coordonnées GPS.
- Visualisation : Affichage des points de concert sur une carte interactive (Tuiles OSM).

### Expérience Utilisateur et Personnalisation (Bonus)
- Système de Favoris : Marquage des groupes préférés avec persistance locale (fichier JSON).
- Import / Export : Partagez votre liste de favoris via des fichiers JSON (Géré dans les paramètres).
- Internationalisation (i18n) : Interface disponible en 4 langues (Français, Anglais, Espagnol, Allemand).
- Thèmes Graphiques : Support natif des modes Clair (Light) et Sombre (Dark).

### Création de Contenu (Bonus)
- Formulaire de création : Possibilité d'ajouter des artistes personnalisés (Nom, Image, Membres, Dates).
- Intégration Mureka : Lien direct pour la génération musicale par IA pour les nouveaux artistes.

---

## 🛠 Stack Technique

- Langage : Go (v1.25.0)
- Framework GUI : Fyne (v2.7.2)
- Architecture : MVC (Model-View-Controller) adapté.
- Données : API RESTful (Source externe) et JSON (Stockage local).
- Services Tiers : Nominatim (OpenStreetMap) pour le géocodage.

---

## 🚀 Prérequis

1. Go : Version 1.21 ou supérieure.
2. Compilateur C (GCC) : Indispensable pour Fyne (liaison CGO).

### Installation des dépendances graphiques :

- Linux (Debian/Ubuntu) : sudo apt-get install golang-go gcc libgl1-mesa-dev xorg-dev
- Windows : Installer TDM-GCC ou Mingw-w64.
- macOS : Installer les Xcode Command Line Tools.

---

## 📦 Installation et Démarrage

1. Cloner le dépôt :
   git clone https://github.com/votre-username/groupie-tracker.git
   cd groupie-tracker

2. Installer les dépendances Go :
   go mod tidy

3. Lancer l'application :
   go run main.go

---

## 📂 Structure du Projet



```text
groupie-tracker/
├── api/                # Gestion des appels API (Fetch, Geocoding)
├── models/             # Structures de données (Artist, Location, Relation)
├── ui/                 # Logique de l'interface utilisateur
│   ├── artists.go         # Liste principale et filtres
│   ├── user_band_form.go  # Formulaire de création
│   ├── settings.go        # Paramètres (Langue, Thème, Export)
│   └── favorites.go       # Gestion des favoris
├── favorites.json      # Sauvegarde des données utilisateur
├── main.go             # Point d'entrée de l'application
├── go.mod              # Dépendances du projet
└── README.md           # Documentation

👥 Auteurs

Projet réalisé par :

Paul

Lina

Aboubakar