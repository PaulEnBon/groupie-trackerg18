Groupie Tracker
Groupie Tracker est une application de bureau développée en Go avec le framework Fyne. Elle permet de visualiser, rechercher et gérer des données sur des artistes et groupes de musique via une API RESTful, tout en offrant des fonctionnalités avancées de personnalisation et de gestion de données utilisateur.

📋 Description
Ce projet va au-delà d'un simple visualiseur de données. Il propose une interface ergonomique (respectant les principes de Shneiderman) pour explorer l'univers musical. L'utilisateur peut non seulement consulter les données de l'API (membres, concerts, dates), mais aussi enrichir l'application en créant ses propres groupes, en gérant ses favoris et en personnalisant l'affichage.

✨ Fonctionnalités Principales
🔍 Recherche et Filtres Avancés
Un système de filtrage puissant pour trouver exactement ce que vous cherchez :

Barre de recherche intelligente : Recherche instantanée par nom d'artiste ou par nom de membre.

Filtres par dates : Sélecteurs de plage pour l'année de création et la date du premier album.

Filtre par membres : Cochez le nombre de membres désiré (ex: duos, trios, groupes de 8+).

Filtre de localisation : Recherchez les groupes passant par une ville spécifique.

Tri dynamique : Ordonnez les résultats par nom, date de création ou date de premier album.

🌟 Gestion des Favoris & Données (Bonus)
Système de favoris : Marquez vos groupes préférés pour les retrouver instantanément.

Import / Export : Sauvegardez votre liste de favoris dans un fichier JSON et importez-la sur une autre machine via le panneau des paramètres.

Persistance : Les données sont sauvegardées localement.

🎸 Création de Groupe Personnalisé (Bonus)
L'application permet d'ajouter vos propres entrées à la liste :

Formulaire complet : Nom, image (upload ou URL), dates, membres.

Gestion des concerts : Ajoutez vos propres dates et lieux.

Liens sociaux : Ajoutez des liens Spotify, YouTube et Deezer.

Intégration AI : Un lien direct vers Mureka pour générer de la musique par IA si vous n'avez pas encore de morceaux !

⚙️ Personnalisation et Paramètres (Bonus)
Internationalisation (i18n) : Interface traduite en 4 langues (Français, Anglais, Espagnol, Allemand).

Thèmes : Basculez entre le mode Clair (Light) et le mode Sombre (Dark).

Affichage : Choix entre une vue Liste détaillée ou une vue Grille plus visuelle.

🗺️ Géolocalisation
Conversion automatique des lieux de concerts en coordonnées géographiques.

Affichage des concerts sur une carte interactive (si implémenté dans la vue détail).

🛠️ Stack Technique
Langage : Go (v1.25.0)

Framework GUI : Fyne (v2.7.2)

Format de données : JSON (API + Sauvegarde locale)

Architecture : Modulaire (api, ui, models)

🚀 Prérequis
Go : Version 1.21 ou supérieure recommandée.

Dépendances C : Un compilateur C (GCC) est requis pour Fyne (pour le rendu graphique OpenGL).

Linux : sudo apt-get install golang-go gcc libgl1-mesa-dev xorg-dev

Windows : TDM-GCC ou Mingw-w64.

macOS : Xcode Command Line Tools.

📦 Installation et Lancement
Cloner le dépôt :

Bash

git clone https://github.com/votre-username/groupie-tracker.git
cd groupie-tracker
Installer les dépendances :

Bash

go mod tidy
Lancer l'application :

Bash

go run main.go
📂 Structure du Projet
main.go : Point d'entrée, initialise l'application et la fenêtre principale.

ui/ : Contient toute la logique de l'interface utilisateur.

artists.go : Liste principale et logique de filtrage.

user_band_form.go : Formulaire de création de groupe.

settings.go : Modale des paramètres (Langue, Thème, Import/Export).

favorites.go : Gestion de la persistance des favoris.

api/ : Gestion des appels vers l'API externe.

models/ : Définition des structures de données (Artist, Location, etc.).

👥 Auteurs
Projet réalisé par Paul, Lina, Aboubakar