package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"go.uber.org/fx"

	"recommendation-system/internal/infrastructure/config"
	"recommendation-system/internal/infrastructure/di"
	"recommendation-system/pkg/algorithm/chain"
	"recommendation-system/pkg/algorithm/strategy"
	"recommendation-system/pkg/datasource"
	"recommendation-system/pkg/plugin"
)