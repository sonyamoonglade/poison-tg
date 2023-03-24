package main

import (
	"context"
	"log"
	"os"

	"github.com/sonyamoonglade/poison-tg/internal/domain"
	"github.com/sonyamoonglade/poison-tg/internal/repositories"
	"github.com/sonyamoonglade/poison-tg/pkg/database"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func main() {
	catalog := []domain.CatalogItem{
		{
			ItemID:         primitive.NewObjectID(),
			ImageURLs:      []string{"https://downloader.disk.yandex.ru/preview/79368452799308513450c3f87c71cb4a82ef82aee82bfeedf45a288ff4275a77/641ca9d2/FpG8BVQ7xplPQuOUv2m3QY_gp0DhcUOLtfLk_OCxEZG7KBA9lRKMVS3FvoNGBC5RX_wQMXeOeuKLtjbstofA5Q%3D%3D?uid=0&filename=2.png&disposition=inline&hash=&limit=0&content_type=image%2Fpng&owner_uid=0&tknv=v2&size=1920x945"},
			Title:          "Jordan",
			Rank:           0,
			AvailableSizes: []string{"42", "43"},
		},
		{
			ItemID:         primitive.NewObjectID(),
			ImageURLs:      []string{"https://downloader.disk.yandex.ru/preview/00b3ac5abb734c67599d98a1358a7300e12288902f7f63b42b7a6b5d524dabe9/641ca9d2/6kXWRw9Yz9JhTxZ4VTxm94_gp0DhcUOLtfLk_OCxEZHJyXbidKaDfCe01aRhbpz5vNTJu8T9bB-1fM4nKGUKOQ%3D%3D?uid=0&filename=3.png&disposition=inline&hash=&limit=0&content_type=image%2Fpng&owner_uid=0&tknv=v2&size=1920x945"},
			Title:          "Air Max",
			Rank:           1,
			AvailableSizes: []string{"42", "44"},
		},
		{
			ItemID:         primitive.NewObjectID(),
			ImageURLs:      []string{"https://downloader.disk.yandex.ru/preview/b43200a3149a9b12eae6c773ec655831635d86ee74421aad02357c2c876bb332/641ca9d2/uBR0gizEoFdYSYKs-HTHGf_jSWrXlsgeYgapGdjpv4i41zlpiqopjTPSbcFJL9ByvQSf8F1QiIuj-jLv1kFulQ%3D%3D?uid=0&filename=4.png&disposition=inline&hash=&limit=0&content_type=image%2Fpng&owner_uid=0&tknv=v2&size=1920x945"},
			Title:          "Adidas",
			Rank:           2,
			AvailableSizes: []string{"44"},
		},
	}

	mongoURI, dbName := os.Getenv("MONGO_URI"), os.Getenv("DB_NAME")
	if mongoURI == "" || dbName == "" {
		log.Fatal("missing envs")
	}
	mongo, err := database.Connect(context.Background(), mongoURI, dbName)
	if err != nil {
		log.Fatal(err)
	}

	catalogRepo := repositories.NewCatalogRepo(mongo.Collection("catalog"), func(items []domain.CatalogItem) {
		return
	})

	for _, item := range catalog {
		if err := catalogRepo.AddItem(context.Background(), item); err != nil {
			log.Fatal(err)
		}
	}
}
