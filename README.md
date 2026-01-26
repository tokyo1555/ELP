LES DOSSIERS:

-filters: Contient les filtres
         *seq.go* = en sequentiel
         *parallel.go* = parallel

-TCP: Connexion serveur-client
      Coté serveur: lancer: *go run server.go parallel.go*
      Coté client: lancer *go run client.go*

-performance: Etude de performance
1. taille de l'image: analyser l’impact de la taille de l’image sur le temps d’exécution des filtres.
   lancer *go run image_size.go parallel.go seq.go*

2. séquentiel VS parallélisme: comparer le temps d’exécution entre les filtres séquentiels et parallèles.
   lancer *go run seq_vs_parallel.go parallel.go seq.go*

3. scaling des goroutines : limite du parallélisme et l’influence du nombre de workers sur les performances.
   lancer *go run scaling_workers.go parallel.go seq.go*
