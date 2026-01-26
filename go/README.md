LES DOSSIERS:

-filters: Contient les filtres *
         *seq.go* = en sequentiel
         *parallel.go* = parallel

-TCP: Connexion serveur-client
      Coté serveur: lancer: *go run server.go parallel.go*
      Coté client: lancer *go run client.go*

-Performance: Etude de performance

1. complexité des filtres: 
   lancer *go run filter_complexity parallel.go seq.go*

2. taille de l'image:

3. séquentiel VS parallélisme:  
   lancer *go run seq_vs_parallel.go parallel.go seq.go*

4. scaling des goroutines : limite du parallélisme
   lancer *go run scaling_workers.go parallel.go seq.go*
